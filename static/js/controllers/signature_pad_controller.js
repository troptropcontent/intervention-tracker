import { Controller } from "../stimulus.js"

export default class extends Controller {
    static targets = ["canvas", "dataInput", "clearButton"]

    connect() {
        this.signaturePad = new SignaturePad(this.canvasTarget, {
            backgroundColor: 'rgb(255, 255, 255)',
            penColor: 'rgb(0, 0, 0)'
        })

        // Update hidden input when signature changes
        this.signaturePad.addEventListener("endStroke", () => {
            this.updateDataInput()
        })

        // Resize canvas to fit container while maintaining aspect ratio
        this.resizeCanvas()
        window.addEventListener('resize', () => this.resizeCanvas())
    }

    disconnect() {
        window.removeEventListener('resize', () => this.resizeCanvas())
    }

    resizeCanvas() {
        const canvas = this.canvasTarget
        const ratio = Math.max(window.devicePixelRatio || 1, 1)

        // Get the container width
        const containerWidth = canvas.parentElement.offsetWidth
        const desiredWidth = Math.min(containerWidth - 32, 800) // Max 800px, with padding
        const desiredHeight = 300

        canvas.width = desiredWidth * ratio
        canvas.height = desiredHeight * ratio
        canvas.style.width = desiredWidth + 'px'
        canvas.style.height = desiredHeight + 'px'
        canvas.getContext('2d').scale(ratio, ratio)

        // Redraw signature if it exists
        if (this.signaturePad && !this.signaturePad.isEmpty()) {
            const data = this.signaturePad.toData()
            this.signaturePad.clear()
            this.signaturePad.fromData(data)
        }
    }

    clear() {
        this.signaturePad.clear()
        this.updateDataInput()
    }

    updateDataInput() {
        if (this.hasDataInputTarget) {
            if (this.signaturePad.isEmpty()) {
                this.dataInputTarget.value = ''
            } else {
                this.dataInputTarget.value = this.signaturePad.toDataURL()
            }
        }
    }
}