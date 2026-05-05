import { useState } from 'react'
import { useAuth } from '../context/AuthContext'
import ImportModal from '../components/ImportModal'
import SignInPrompt from '../components/SignInPrompt'

export function useUploadDeck(onImported) {
  const { user } = useAuth()
  const [mode, setMode] = useState(null)

  const open = () => setMode(user ? 'import' : 'signin')
  const close = () => setMode(null)

  const modal = (
    <>
      {mode === 'import' && (
        <ImportModal onClose={close} onImported={() => { close(); onImported?.() }} />
      )}
      {mode === 'signin' && <SignInPrompt onClose={close} />}
    </>
  )

  return { open, modal, isAuthed: !!user }
}
