var acToken = document.querySelector('Authorization').value;
if (acToken !== null) {
    localStorage.setItem('act', acToken);
} else {
    var storedValue = localStorage.getItem('act');
    document.getElementById('Authorization').innerText = storedValue;
}
