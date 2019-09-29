let moodInput = document.querySelector("#mood");
let moodGradient = document.querySelector("#mood-gradient");

moodGradient.classList.remove("hidden");

moodGradient.addEventListener("click", function(ev) {
	moodInput.value = ev.clientX / document.body.clientWidth;
});

