let moodInput = document.querySelector("#entry-value");
let moodGradient = document.querySelector("#mood-gradient");

moodGradient.classList.remove("hidden");

moodGradient.addEventListener("click", function(ev) {
	moodInput.value = ev.clientX / document.body.clientWidth;
});

