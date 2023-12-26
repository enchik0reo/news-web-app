const createForm = document.getElementById('createForm');
if (typeof createForm !== 'undefined' && createForm !== null) {
   createForm.addEventListener('submit', function(event) {
   event.preventDefault();

   const link = document.getElementById('link').value;
   const content = document.getElementById('content').value;

   if (typeof link !== 'undefined' && link !== null &&
   typeof content !== 'undefined' && content !== null) {
    fetch('/article/suggest', {
      method: 'POST',
      body: JSON.stringify({ link: link, content: content }),
      headers: {
        'Content-Type': 'application/json'
      }
    })
    .then(response => response.json())
    .then(data => {
    
      console.log('Article successfully suggested.');

      document.body.innerHTML = data; // заменить содержимое страницы загруженным HTML
    })
    .catch(error => {
      console.error('Произошла ошибка:', error);
    });
   } else {
    console.log('Переменные link и/или description не определены.');
   }
 });
}