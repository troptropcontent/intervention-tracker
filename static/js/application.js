import "./htmx.min.js"
import { Application } from "./stimulus.js"
import QrCodeScannerController from "./controllers/qr_code_scanner_controller.js"

window.Stimulus = Application.start()
Stimulus.register("qr-code-scanner", QrCodeScannerController)