document.addEventListener("DOMContentLoaded", function() {
    const token = localStorage.getItem('token');
     if (token) {
        console.log('Переменная token найдена')
        var currentUrl = window.location.href;
      fetch("/refresh", {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + token 
        }
      })
      .then(response => response.json())
      .then(data => {
        const token = data.AccessToken;
  
        if (typeof token !== 'undefined' && token !== null) {
             // Запись токена в localStorage
          localStorage.setItem('token', token);

          console.log('Токен успешно сохранен в localStorage после refresh.');

          window.location.href = currentUrl;
        }
      })
      .catch(error => {
        console.log('Response не в формате json:', error);
      });
    } else {
      console.log('Переменная token не найдена в localstorage');
    }
  });