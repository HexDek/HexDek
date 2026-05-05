import { useState } from 'react'
import ImportModal from '../components/ImportModal'

export function useUploadDeck(onImported) {
  const [open, setOpen] = useState(false)

  const modal = open ? (
    <ImportModal
      onClose={() => setOpen(false)}
      onImported={() => { setOpen(false); onImported?.() }}
    />
  ) : null

  return {
    open: () => setOpen(true),
    modal,
    isOpen: open,
  }
}
