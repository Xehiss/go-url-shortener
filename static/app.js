document.getElementById('url-form').addEventListener('submit', function(e) {
    e.preventDefault();

    var url = document.getElementById('url').value;

    fetch('http://localhost:8080/create?url=' + encodeURIComponent(url), {
        method: 'GET',
    })
    .then(response => response.text())
    .then(data => {
        document.getElementById('result').style.display = 'block';
        document.getElementById('short-url').href = data;
        document.getElementById('short-url').textContent = data;
        generateQR(data);
    })
    .catch((error) => {
        console.error('Error:', error);
    });
});

function generateQR(shortUrl) {
    // Clear out the QR code div first, in case a previous QR code was generated
    var qrDiv = document.getElementById("qrcode");
    qrDiv.innerHTML = "";
    
    var qrcode = new QRCode(qrDiv, {
        text: shortUrl,
        width: 128,
        height: 128,
        colorDark : "#000000",
        colorLight : "#ffffff",
        correctLevel : QRCode.CorrectLevel.H
    });
}

