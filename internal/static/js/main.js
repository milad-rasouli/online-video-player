
window.onload = ()=>{
}

function sendAccessTokenRequest() {
  fetch('/auth/update-token', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        'JWT-Token':'access',
      },
  })
  .then(response => {
      if (response === null) {
          console.log('No response from server');
          return;
      }
      return true;
  })
  .catch(error => {
      console.log('Error refreshing token:', error);
  });
}
sendAccessTokenRequest();
setInterval(sendAccessTokenRequest, 6000);