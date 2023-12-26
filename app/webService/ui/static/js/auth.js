document.addEventListener("DOMContentLoaded", function() {
  const token = localStorage.getItem('token');
   if (token) {
      console.log('Переменная token найдена')
     var currentUrl = window.location.href;
    fetch(currentUrl, {
      headers: {
        'Authorization': 'Bearer ' + token 
      }
      })
    .then(response => response.text())
    .then(data => {
      // удалить строку с заголовком 'Authorization' из HTML-страницы
      data = data.replace(`{"AccessToken":"Bearer ${token}"}`, ''); 
      document.body.innerHTML = data; // заменить содержимое страницы загруженным HTML
    })
    .catch(error => {
      console.error('Произошла ошибка при загрузке страницы:', error);
    });
  } else {
    console.log('Переменная token не найдена');
  }
});