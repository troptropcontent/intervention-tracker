import { Controller } from "../stimulus.js"

export default class extends Controller {
    static targets = ["modal", "loading", "status", "error", "reader", "errorMessage"]
    static values = { portalId: String }

    connect() {
        this.scanner = null
        this.isScanning = false
        this.qrUuidRegex = /\/qr_codes\/([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})/i
    }

    disconnect() {
        this.cleanup()
    }

    openQRScanner() {
        if (!this.modalTarget) return

        this.showModal()
        this.initializeScanner()
    }

    closeQRScanner() {
        this.hideModal()
        this.cleanup()
    }

    showModal() {
        this.modalTarget.style.display = 'block'
        document.body.style.overflow = 'hidden'
    }

    hideModal() {
        this.modalTarget.style.display = 'none'
        document.body.style.overflow = 'auto'
    }

    async initializeScanner() {
        if (this.isScanning) return

        this.showLoading()
        this.hideError()

        try {
            this.validateDependencies()
            await this.setupCamera()
        } catch (error) {
            this.showError(error.message)
        }
    }

    validateDependencies() {
        if (typeof Html5Qrcode === 'undefined') {
            throw new Error("La bibliothèque de scan QR n'a pas pu être chargée")
        }
    }

    async setupCamera() {
        try {
            const devices = await Html5Qrcode.getCameras()
            
            if (!devices || devices.length === 0) {
                throw new Error("Aucune caméra détectée sur cet appareil")
            }

            const cameraId = devices[devices.length - 1].id
            await this.startScanning(cameraId)
            
        } catch (error) {
            throw new Error(`Erreur lors de l'accès aux caméras: ${error}`)
        }
    }

    async startScanning(cameraId) {
        this.scanner = new Html5Qrcode(this.readerTarget.id)
        
        const config = {
            fps: 10,
            qrbox: { width: 250, height: 250 },
            aspectRatio: 1.0
        }

        try {
            await this.scanner.start(cameraId, config, (decodedText) => {
                this.handleQRCodeDetected(decodedText)
            })

            this.isScanning = true
            this.hideLoading()
            this.showStatus("Pointez la caméra vers le QR code")
            
        } catch (error) {
            throw new Error(`Erreur lors du démarrage de la caméra: ${error}`)
        }
    }

    async handleQRCodeDetected(decodedText) {
        const match = decodedText.match(this.qrUuidRegex)
        
        if (!match) return

        const qrCodeId = match[1]
        
        this.hideReader()
        this.showStatus("QR code détecté, association en cours...", "text-green-600")
        
        await this.stopScanning()
        await this.processQRCode(qrCodeId)
    }

    async processQRCode(qrCodeId) {
        try {
            if (!this.portalIdValue) {
                throw new Error("ID du portail non trouvé")
            }

            const response = await fetch(`/admin/portals/${this.portalIdValue}/qr-code/associate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    qr_code_uuid: qrCodeId
                })
            })
            
            if (response.ok) {
                const htmlContent = await response.text()
                const targetElement = document.getElementById('qr_code_association_section')
                if (targetElement) {
                    targetElement.innerHTML = htmlContent
                }
                this.showStatus("Association réussie!", "text-green-600")
                setTimeout(() => {
                    this.closeQRScanner()
                }, 1500)
            } else {
                const errorData = await response.text()
                throw new Error(errorData || "Erreur lors de l'association du QR code")
            }
            
        } catch (error) {
            this.showError(`Erreur API: ${error.message}`)
            this.showReader()
            await this.restartScanning()
        }
    }

    async restartScanning() {
        if (this.scanner && !this.isScanning) {
            try {
                const devices = await Html5Qrcode.getCameras()
                const cameraId = devices[devices.length - 1].id
                await this.startScanning(cameraId)
            } catch (error) {
                this.showError(`Impossible de redémarrer le scanner: ${error.message}`)
            }
        }
    }

    async stopScanning() {
        if (this.scanner && this.isScanning) {
            try {
                await this.scanner.stop()
                this.isScanning = false
            } catch (error) {
                console.error("Erreur lors de l'arrêt du scanner:", error)
            }
        }
    }

    cleanup() {
        this.stopScanning()
        this.scanner = null
    }

    showLoading() {
        if (this.hasLoadingTarget) {
            this.loadingTarget.style.display = 'block'
        }
    }

    hideLoading() {
        if (this.hasLoadingTarget) {
            this.loadingTarget.style.display = 'none'
        }
    }

    showStatus(message, className = "text-gray-600") {
        if (this.hasStatusTarget) {
            this.statusTarget.innerHTML = `<p class="${className} font-medium">${message}</p>`
            this.statusTarget.style.display = 'block'
        }
    }

    hideStatus() {
        if (this.hasStatusTarget) {
            this.statusTarget.style.display = 'none'
        }
    }

    showError(message) {
        this.hideStatus()
        
        if (this.hasErrorMessageTarget) {
            this.errorMessageTarget.textContent = message
        }
        if (this.hasErrorTarget) {
            this.errorTarget.style.display = 'block'
        }
    }

    hideError() {
        if (this.hasErrorTarget) {
            this.errorTarget.style.display = 'none'
        }
    }

    showReader() {
        if (this.hasReaderTarget) {
            this.readerTarget.style.display = 'block'
        }
    }

    hideReader() {
        if (this.hasReaderTarget) {
            this.readerTarget.style.display = 'none'
        }
    }
}