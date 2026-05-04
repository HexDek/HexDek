#!/usr/bin/env python3
"""
Train the Level 5 neural position evaluator from grinder data.

Reads training samples from data/training/samples.jsonl (produced by the
Go grinder), trains a deep MLP (92→256→128→64→1), and exports the model
to data/training/model.json for Go inference.

Requirements: pip install torch  (CUDA optional, auto-detected)

Usage:
    python scripts/train_neural_eval.py [--samples data/training/samples.jsonl]
                                        [--output data/training/model.json]
                                        [--epochs 200]
                                        [--lr 0.001]
                                        [--batch 512]
                                        [--patience 20]
                                        [--checkpoint-dir data/training/checkpoints]
"""

import argparse
import json
import os
import sys
import time

import torch
import torch.nn as nn
from torch.utils.data import DataLoader, TensorDataset

STATE_DIM = 92  # must match hat.StateVectorDim


def load_samples(path: str) -> tuple[torch.Tensor, torch.Tensor, torch.Tensor]:
    """Load JSONL samples. Returns (states, placements, pivot_weights).

    Pivot weights are derived from Tesla causal graph data: samples near
    the causal pivot get higher weight (more training signal), samples
    far from the pivot get lower weight.
    """
    states, placements, pivot_weights = [], [], []
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            sample = json.loads(line)
            s = sample["state"]
            if len(s) != STATE_DIM:
                continue
            states.append(s)
            placements.append(sample["placement"])
            # Tesla pivot distance: 0.0 = at pivot (high weight), 1.0 = far (low weight).
            pd = sample.get("pivot_distance", 0.5)
            pivot_weights.append(1.0 + (1.0 - pd))  # range [1.0, 2.0]
    X = torch.tensor(states, dtype=torch.float32)
    y = torch.tensor(placements, dtype=torch.float32).unsqueeze(1)
    w = torch.tensor(pivot_weights, dtype=torch.float32).unsqueeze(1)
    return X, y, w


class PositionEvaluator(nn.Module):
    def __init__(self, layers: list[int] | None = None):
        super().__init__()
        if layers is None:
            layers = [STATE_DIM, 256, 128, 64, 1]
        blocks = []
        for i in range(len(layers) - 1):
            blocks.append(nn.Linear(layers[i], layers[i + 1]))
            if i < len(layers) - 2:
                blocks.append(nn.ReLU())
                blocks.append(nn.Dropout(0.1))
        blocks.append(nn.Sigmoid())
        self.net = nn.Sequential(*blocks)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        return self.net(x)


def export_model(model: PositionEvaluator, path: str, meta: dict) -> None:
    """Export trained model to JSON for Go inference."""
    layers_json = []
    for module in model.net:
        if isinstance(module, nn.Linear):
            layers_json.append({
                "weights": module.weight.detach().cpu().tolist(),
                "biases": module.bias.detach().cpu().tolist(),
            })

    out = {"arch": "mlp", "layers": layers_json, "meta": meta}
    os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
    with open(path, "w") as f:
        json.dump(out, f)


def main():
    parser = argparse.ArgumentParser(description="Train HexDek neural evaluator")
    parser.add_argument("--samples", default="data/training/samples.jsonl")
    parser.add_argument("--output", default="data/training/model.json")
    parser.add_argument("--epochs", type=int, default=200)
    parser.add_argument("--lr", type=float, default=0.001)
    parser.add_argument("--batch", type=int, default=512)
    parser.add_argument("--patience", type=int, default=20)
    parser.add_argument("--checkpoint-dir", default="data/training/checkpoints")
    args = parser.parse_args()

    if not os.path.exists(args.samples):
        print(f"No training data at {args.samples}")
        print("Run the grinder first to accumulate samples.")
        sys.exit(1)

    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    print(f"Device: {device}", end="")
    if device.type == "cuda":
        print(f" ({torch.cuda.get_device_name(0)}, {torch.cuda.get_device_properties(0).total_memory / 1e9:.1f} GB)")
    else:
        print()

    print(f"Loading samples from {args.samples}...")
    X, y, w = load_samples(args.samples)
    n = len(X)
    print(f"  {n} samples, {STATE_DIM} features")

    if n < 100:
        print("Need at least 100 samples to train. Run more grinder games.")
        sys.exit(1)

    # 90/10 train/val split.
    perm = torch.randperm(n, generator=torch.Generator().manual_seed(42))
    n_val = max(1, n // 10)
    X_val, y_val = X[perm[:n_val]].to(device), y[perm[:n_val]].to(device)
    X_train, y_train = X[perm[n_val:]], y[perm[n_val:]]
    w_train = w[perm[n_val:]]
    print(f"  train={len(X_train)}, val={len(X_val)}")

    train_ds = TensorDataset(X_train, y_train, w_train)
    train_dl = DataLoader(train_ds, batch_size=args.batch, shuffle=True,
                          pin_memory=(device.type == "cuda"), num_workers=0)

    arch = [STATE_DIM, 256, 128, 64, 1]
    model = PositionEvaluator(arch).to(device)
    total_params = sum(p.numel() for p in model.parameters())
    print(f"  Architecture: {' → '.join(map(str, arch))}")
    print(f"  Parameters: {total_params:,}")

    optimizer = torch.optim.Adam(model.parameters(), lr=args.lr, weight_decay=1e-5)
    scheduler = torch.optim.lr_scheduler.ReduceLROnPlateau(
        optimizer, mode="min", factor=0.5, patience=8, min_lr=1e-6)
    criterion = nn.MSELoss()
    criterion_none = nn.MSELoss(reduction="none")

    os.makedirs(args.checkpoint_dir, exist_ok=True)
    best_val_loss = float("inf")
    epochs_no_improve = 0
    best_epoch = 0
    t0 = time.time()

    print(f"\nTraining (epochs={args.epochs}, lr={args.lr}, patience={args.patience})...")

    for epoch in range(args.epochs):
        model.train()
        train_loss = 0.0
        n_batches = 0
        for Xb, yb, wb in train_dl:
            Xb, yb, wb = Xb.to(device), yb.to(device), wb.to(device)
            pred = model(Xb)
            loss = (criterion_none(pred, yb) * wb).mean()
            optimizer.zero_grad()
            loss.backward()
            optimizer.step()
            train_loss += loss.item()
            n_batches += 1
        avg_train = train_loss / max(n_batches, 1)

        model.eval()
        with torch.no_grad():
            val_pred = model(X_val)
            val_loss = criterion(val_pred, y_val).item()

        scheduler.step(val_loss)
        current_lr = optimizer.param_groups[0]["lr"]

        if epoch % 10 == 0 or epoch == args.epochs - 1 or val_loss < best_val_loss:
            elapsed = time.time() - t0
            print(f"  epoch {epoch:4d}/{args.epochs}: "
                  f"train={avg_train:.6f}  val={val_loss:.6f}  "
                  f"lr={current_lr:.2e}  [{elapsed:.0f}s]")

        if val_loss < best_val_loss:
            best_val_loss = val_loss
            best_epoch = epoch
            epochs_no_improve = 0
            ckpt_path = os.path.join(args.checkpoint_dir, "best.pt")
            torch.save(model.state_dict(), ckpt_path)
        else:
            epochs_no_improve += 1
            if epochs_no_improve >= args.patience:
                print(f"\n  Early stopping at epoch {epoch} (best was {best_epoch})")
                break

    # Load best checkpoint.
    best_ckpt = os.path.join(args.checkpoint_dir, "best.pt")
    if os.path.exists(best_ckpt):
        model.load_state_dict(torch.load(best_ckpt, map_location=device, weights_only=True))

    # Final validation.
    model.eval()
    with torch.no_grad():
        val_pred = model(X_val).squeeze()
        val_targets = y_val.squeeze()
        final_mse = torch.mean((val_pred - val_targets) ** 2).item()
        if len(val_targets) > 1:
            vp = val_pred - val_pred.mean()
            vt = val_targets - val_targets.mean()
            corr = (torch.sum(vp * vt) /
                    (torch.sqrt(torch.sum(vp ** 2)) * torch.sqrt(torch.sum(vt ** 2)) + 1e-8))
            val_corr = corr.item()
        else:
            val_corr = 0.0

    elapsed = time.time() - t0
    print(f"\nValidation: MSE={final_mse:.6f}, correlation={val_corr:.4f}")
    print(f"Training time: {elapsed:.1f}s ({elapsed / 60:.1f}min)")

    meta = {
        "input_dim": STATE_DIM,
        "arch": arch,
        "epochs_trained": best_epoch + 1,
        "best_val_loss": best_val_loss,
        "val_mse": final_mse,
        "val_corr": val_corr,
        "n_samples": n,
        "n_params": total_params,
        "device": str(device),
    }
    export_model(model, args.output, meta)
    size_kb = os.path.getsize(args.output) / 1024
    print(f"Model saved to {args.output} ({size_kb:.1f} KB)")


if __name__ == "__main__":
    main()
