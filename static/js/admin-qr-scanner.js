// QR Code Scanner for Admin Portal Association
let html5QrCode = null;
let isScanning = false;
let scannerModal = null;
let currentPortalId = null;

function openQRScanner(buttonElement) {
    // Get portal ID from the button's data attribute
    currentPortalId = buttonElement.dataset.portalId;
    
    const modal = document.getElementById('qr-scanner-modal');
    if (modal) {
        modal.style.display = 'block';
        document.body.style.overflow = 'hidden'; // Prevent background scrolling
        initAdminQRScanner();
    }
}

function closeQRScanner() {
    const modal = document.getElementById('qr-scanner-modal');
    if (modal) {
        modal.style.display = 'none';
        document.body.style.overflow = 'auto'; // Restore scrolling
        stopAdminScanning();
    }
}

function initAdminQRScanner() {
    const readerId = "admin-qr-reader";
    const loadingEl = document.getElementById("scanner-loading");
    const statusEl = document.getElementById("scanner-status");
    const errorEl = document.getElementById("scanner-error");

    // Check if Html5Qrcode is available
    if (typeof Html5Qrcode === 'undefined') {
        showScannerError("La bibliothèque de scan QR n'a pas pu être chargée");
        return;
    }

    html5QrCode = new Html5Qrcode(readerId);

    // Scanner configuration
    const config = {
        fps: 10,
        qrbox: { width: 250, height: 250 },
        aspectRatio: 1.0
    };

    // Success callback
    function onScanSuccess(decodedText, decodedResult) {
        if (isScanning) {
            isScanning = false;
            console.log(`QR Code detected: ${decodedText}`);
            
            // Stop scanner
            stopAdminScanning();
            
            // Extract UUID from scanned URL
            const uuid = extractQRCodeUUID(decodedText);
            
            if (uuid) {
                // Show processing message
                showScannerSuccess(`QR Code détecté: ${uuid.substring(0, 8)}... Association en cours...`);
                
                // Call the backend API directly
                associateQRCode(uuid);
            } else {
                showScannerError("QR code invalide - impossible d'extraire l'UUID");
                // Restart scanner after error
                setTimeout(() => {
                    if (document.getElementById('qr-scanner-modal').style.display === 'block') {
                        initAdminQRScanner();
                    }
                }, 3000);
            }
        }
    }

    // Error callback (don't show continuous scan errors)
    function onScanFailure(error) {
        // Ignore continuous scan failures - they're normal
    }

    // Start scanner with preferred camera
    Html5Qrcode.getCameras().then(devices => {
        
        if (loadingEl) loadingEl.style.display = 'none';
        
        if (devices && devices.length) {
            // Prefer back camera
            let cameraId = devices.slice(-1)[0].id;
            const backCamera = devices.find(device => 
                device.label.toLowerCase().includes('back') || 
                device.label.toLowerCase().includes('rear') ||
                device.label.toLowerCase().includes('environment')
            );
            
            if (backCamera) {
                cameraId = backCamera.id;
            }

            // Start scanning
            html5QrCode.start(
                cameraId, 
                config,
                onScanSuccess,
                onScanFailure
            ).then(() => {
                isScanning = true;
                if (statusEl) {
                    statusEl.innerHTML = '<p class="text-gray-600">Pointez la caméra vers le QR code</p>';
                }
            }).catch(err => {
                showScannerError(`Erreur lors du démarrage de la caméra: ${err}`);
            });
        } else {
            showScannerError("Aucune caméra détectée sur cet appareil");
        }
    }).catch(err => {
        if (loadingEl) loadingEl.style.display = 'none';
        showScannerError(`Erreur lors de l'accès aux caméras: ${err}`);
    });
}

function stopAdminScanning() {
    if (html5QrCode && isScanning) {
        html5QrCode.stop().then(() => {
            console.log("Admin scanner stopped");
            isScanning = false;
        }).catch(err => {
            console.error("Error stopping admin scanner:", err);
        });
    }
}

function extractQRCodeUUID(url) {
    try {
        // Extract UUID from /qr_codes/ URLs
        let uuidRegex = /\/qr_codes\/([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})/i;
        let match = url.match(uuidRegex);
        
        if (match) {
            return match[1];
        }
        
        return null;
    } catch (error) {
        console.error('Error extracting UUID:', error);
        return null;
    }
}

function showScannerError(message) {
    const errorEl = document.getElementById("scanner-error");
    const errorMessage = document.getElementById("scanner-error-message");
    const statusEl = document.getElementById("scanner-status");
    
    if (statusEl) statusEl.style.display = 'none';
    if (errorMessage) errorMessage.textContent = message;
    if (errorEl) errorEl.style.display = 'block';
    
    console.error("Admin QR Scanner Error:", message);
}

function showScannerSuccess(message) {
    const errorEl = document.getElementById("scanner-error");
    const statusEl = document.getElementById("scanner-status");
    
    if (errorEl) errorEl.style.display = 'none';
    if (statusEl) {
        statusEl.innerHTML = `<p class="text-green-600 font-medium">${message}</p>`;
        statusEl.style.display = 'block';
    }
}

// Clean up on page unload
window.addEventListener('beforeunload', function() {
    stopAdminScanning();
});

// Handle visibility change (tab switching)
document.addEventListener('visibilitychange', function() {
    if (document.hidden && isScanning) {
        stopAdminScanning();
    }
});

// Close modal when clicking outside
window.addEventListener('click', function(event) {
    const modal = document.getElementById('qr-scanner-modal');
    if (event.target === modal) {
        closeQRScanner();
    }
});