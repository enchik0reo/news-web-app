var acToken = document.getElementsByName('Authorization')[0].value;
if (acToken !== null) {
    window.localStorage.setItem('act', acToken);
} else {
    var myToken = window.localStorage.getItem('act');
    var myAnswer = document.getElementById('token');
    myAnswer.innerHTML = myToken;
}
// возможно хранить в <input type="hidden" name="token" class="token-control"></input> или div вместо input