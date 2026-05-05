import { useNavigate } from 'react-router-dom'
import ImportModal from '../components/ImportModal'

// Import — full-page deck-import route at /import.
//
// This route now renders the unified ImportModal as a full-page overlay.
// Users who navigate directly to /import (bookmarks, links) get the same
// unified experience as those who trigger the modal from Dashboard or DeckList.
// On close (cancel/escape), we navigate back; on success the modal handles
// the redirect to the new deck's archive page.

export default function Import() {
  const navigate = useNavigate()

  const handleClose = () => {
    // Go back if there's history, otherwise go to decks list
    if (window.history.length > 1) {
      navigate(-1)
    } else {
      navigate('/decks')
    }
  }

  return (
    <ImportModal
      onClose={handleClose}
      onImported={() => {}}
    />
  )
}
