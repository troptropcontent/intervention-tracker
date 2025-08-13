// QR Code Scanner functionality
let html5QrCode = null;
let isScanning = false;

function initQRScanner() {
    const readerId = "reader";
    const loadingEl = document.getElementById("loading");
    const statusEl = document.getElementById("status");
    const errorEl = document.getElementById("error");
    const successEl = document.getElementById("success");

    // Import html5-qrcode from node_modules
    if (typeof Html5Qrcode === 'undefined') {
        showError("La bibliothèque de scan QR n'a pas pu être chargée");
        return;
    }

    html5QrCode = new Html5Qrcode(readerId);

    // Configuration du scanner
    const config = {
        fps: 10,
        qrbox: { width: 250, height: 250 },
        aspectRatio: 1.0
    };

    // Fonction appelée lors du scan réussi
    function onScanSuccess(decodedText, decodedResult) {
        if (isScanning) {
            isScanning = false;
            console.log(`QR Code détecté: ${decodedText}`);
            
            // Arrêter le scanner
            stopScanning();
            
            // Afficher le message de succès
            showSuccess();
            
            // Traiter l'URL scannée
            processScannedURL(decodedText);
        }
    }

    // Fonction appelée en cas d'erreur de scan (ne pas afficher ces erreurs)
    function onScanFailure(error) {
        // Ne pas afficher les erreurs de scan continu - c'est normal
        // console.warn(`Erreur de scan: ${error}`);
    }

    // Démarrer le scanner avec la caméra arrière si disponible
    Html5Qrcode.getCameras().then(devices => {
        loadingEl.style.display = 'none';
        
        if (devices && devices.length) {
            // Préférer la caméra arrière
            let cameraId = devices[0].id;
            const backCamera = devices.find(device => 
                device.label.toLowerCase().includes('back') || 
                device.label.toLowerCase().includes('rear') ||
                device.label.toLowerCase().includes('environment')
            );
            
            if (backCamera) {
                cameraId = backCamera.id;
            }

            // Démarrer le scan
            html5QrCode.start(
                cameraId, 
                config,
                onScanSuccess,
                onScanFailure
            ).then(() => {
                isScanning = true;
                statusEl.innerHTML = '<p class="text-gray-600">Caméra activée - Pointez vers un QR code</p>';
            }).catch(err => {
                showError(`Erreur lors du démarrage de la caméra: ${err}`);
            });
        } else {
            showError("Aucune caméra détectée sur cet appareil");
        }
    }).catch(err => {
        loadingEl.style.display = 'none';
        showError(`Erreur lors de l'accès aux caméras: ${err}`);
    });
}

function stopScanning() {
    if (html5QrCode && isScanning) {
        html5QrCode.stop().then(() => {
            console.log("Scanner arrêté");
        }).catch(err => {
            console.error("Erreur lors de l'arrêt du scanner:", err);
        });
    }
}

function processScannedURL(scannedText) {
    console.log("Traitement de l'URL scannée:", scannedText);
    
    // Extraire l'UUID de l'URL
    const uuid = extractUUIDFromURL(scannedText);
    
    if (uuid) {
        // Redirection vers la page du portail
        const portalURL = `/portals/${uuid}`;
        console.log("Redirection vers:", portalURL);
        
        setTimeout(() => {
            window.location.href = portalURL;
        }, 1500); // Délai pour voir le message de succès
    } else {
        showError("QR code invalide - l'URL ne correspond pas à un portail");
        // Relancer le scanner après quelques secondes
        setTimeout(() => {
            location.reload();
        }, 3000);
    }
}

function extractUUIDFromURL(url) {
    // Regex pour extraire un UUID d'une URL
    const uuidRegex = /\/portals\/([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})/i;
    const match = url.match(uuidRegex);
    return match ? match[1] : null;
}

function setupManualInput() {
    const manualInput = document.getElementById("manual-input");
    const manualSubmit = document.getElementById("manual-submit");

    manualSubmit.addEventListener('click', function() {
        const url = manualInput.value.trim();
        if (url) {
            processScannedURL(url);
        } else {
            showError("Veuillez saisir une URL valide");
        }
    });

    // Permettre la soumission avec Entrée
    manualInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            manualSubmit.click();
        }
    });
}

function showError(message) {
    const errorEl = document.getElementById("error");
    const errorMessage = document.getElementById("error-message");
    const successEl = document.getElementById("success");
    
    successEl.style.display = 'none';
    errorMessage.textContent = message;
    errorEl.style.display = 'block';
    
    console.error("Erreur QR Scanner:", message);
}

function showSuccess() {
    const errorEl = document.getElementById("error");
    const successEl = document.getElementById("success");
    
    errorEl.style.display = 'none';
    successEl.style.display = 'block';
}

// Nettoyer lors de la fermeture de la page
window.addEventListener('beforeunload', function() {
    stopScanning();
});

// Gérer la perte de focus (changement d'onglet)
document.addEventListener('visibilitychange', function() {
    if (document.hidden) {
        stopScanning();
    } else if (html5QrCode && !isScanning) {
        // Recharger la page pour relancer le scanner
        location.reload();
    }
});