import "./htmx.min.js"
import { Application } from "./stimulus.js"
import QrCodeScannerController from "./controllers/qr_code_scanner_controller.js"
import LogoutController from "./controllers/logout_controller.js"
import SignaturePadController from "./controllers/signature_pad_controller.js"
import InterventionPhotoInputsController from "./controllers/intervention_photo_inputs_controller.js"

window.Stimulus = Application.start()
Stimulus.register("qr-code-scanner", QrCodeScannerController)
Stimulus.register("logout", LogoutController)
Stimulus.register("signature-pad", SignaturePadController)
Stimulus.register("intervention-photo-inputs", InterventionPhotoInputsController)