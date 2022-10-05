'use strict';

const generateButton = document.querySelector(".generateButton");

generateButton.addEventListener('click', () => {

  const min = parseInt(document.querySelector(".min").value);
  const max = parseInt(document.querySelector(".max").value);

  if (min == min && max == max && min < max) {
    let randomNumber = Math.floor(Math.random()*(max-min+1))+min;
    const randomNumberResult = document.querySelector(".randomNumberResult");
    randomNumberResult.innerHTML = randomNumber;
  }
});
