import { Controller } from "../stimulus.js"

export default class extends Controller {
  static targets = ["photoInputGroupTemplate", "photoInputGroup", "photoInputGroups", "photoInputGroupNameInput"]

  addNewPhotoInputGroup() {
    const template = this.photoInputGroupTemplateTarget
    const container = this.photoInputGroupsTarget
    const currentIndex = this.photoInputGroupTargets.length

    const templateHtml = template.innerHTML.replace(/\{\{index\}\}/g, currentIndex)
    const wrapper = document.createElement('div')
    wrapper.innerHTML = templateHtml

    container.appendChild(wrapper.firstElementChild)
  }

  changeNameInputValue(e) {
    const file = e.target.files?.[0]
    if (!file) return

    const match = e.target.name.match(/photos\[(?<index>\d+)\]file/)
    const index = match?.groups?.index
    if (!index) return

    const nameInput = this.photoInputGroupNameInputTargets.find(
      input => input.name === `photos[${index}]name`
    )

    if (nameInput) {
      nameInput.value = file.name
    }
  }
}