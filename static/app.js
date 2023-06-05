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
    })
    .catch((error) => {
        console.error('Error:', error);
    });
});

