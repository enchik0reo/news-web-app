 const loginForm = document.getElementById('loginForm');
 if (typeof loginForm !== 'undefined' && loginForm !== null) {
  loginForm.addEventListener('submit', function(event) {
    event.preventDefault();
 
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
 
    if (typeof email !== 'undefined' && email !== null &&
    typeof password !== 'undefined' && password !== null) {
     fetch('/login', {
       method: 'POST',
       body: JSON.stringify({ email: email, password: password }),
       headers: {
         'Content-Type': 'application/json'
       }
     })
     .then(response => response.json())
     .then(data => {
       // Получение токена из ответа сервера
       const token = data.AccessToken;
  
       // Запись токена в localStorage
       localStorage.setItem('token', token);
  
       console.log('Токен успешно сохранен в localStorage.');

       alert('You have successfully logged in!');
       // Перенаправление на другую страницу
       window.location.href = "/";
     })
     .catch(error => {
       alert('Incorrect e-mail or password.');
       console.error('Произошла ошибка:', error);
     });
    } else {
     alert('Incorrect e-mail or password.');
     console.log('Переменные email и/или password не определены.');
    }
  });
 }