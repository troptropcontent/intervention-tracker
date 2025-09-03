import { Controller } from "../stimulus.js"

export default class extends Controller {
  async logout() {
    try {
      const response = await fetch('/logout', {
        method: 'POST',
        headers: {
          'X-Requested-With': 'XMLHttpRequest'
        }
      })
      
      if (response.ok) {
        window.location.href = '/login'
      } else {
        console.error('Logout failed')
      }
    } catch (error) {
      console.error('Logout error:', error)
      window.location.href = '/login'
    }
  }
}