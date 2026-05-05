import { useState } from 'react'
import { useAuth } from '../context/AuthContext'
import ImportModal from '../components/ImportModal'
import AuthPrompt from '../components/AuthPrompt'

export function useUploadDeck(onImported) {
  const { user } = useAuth()
  const [mode, setMode] = useState(null)

  const open = () => setMode(user ? 'import' : 'auth')
  const close = () => setMode(null)

  const modal = (
    <>
      {mode === 'import' && (
        <ImportModal onClose={close} onImported={() => { close(); onImported?.() }} />
      )}
      {mode === 'auth' && <AuthPrompt onClose={close} action="upload deck" />}
    </>
  )

  return { open, modal, isAuthed: !!user }
}
