function removeToken() {
    localStorage.removeItem('token');
    alert('You have successfully logged out!');
    window.location.href = "/";
}