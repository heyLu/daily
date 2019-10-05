let additionalFieldsContainer = document.querySelector("#additional-fields");
let addField = document.querySelector("#add-field");

addField.addEventListener("click", function(ev) {
	let fieldContainer = document.createElement("div");
	fieldContainer.classList.add("field");

	let fieldName = document.createElement("input");
	fieldName.type = "text";
	fieldName.placeholder = "field name";
	fieldContainer.appendChild(fieldName);

	let fieldValue = document.createElement("input");
	fieldValue.type = "text";
	fieldValue.placeholder = "field value";
	fieldContainer.appendChild(fieldValue);

	fieldName.addEventListener("input", function(ev) {
		fieldValue.name = fieldName.value;
	});

	additionalFieldsContainer.appendChild(fieldContainer);

	fieldName.focus();

	// prevent form submit
	ev.preventDefault();
});

// connect existing field keys to field values
let additionalFields = additionalFieldsContainer.querySelectorAll(".field");
for (let i = 0; i < additionalFields.length; i++) {
	let fieldName = additionalFields[i].querySelector(".field-key");
	let fieldValue = additionalFields[i].querySelector(".field-value");

	fieldName.addEventListener("input", function(ev) {
		fieldValue.name = fieldName.value;
	});
}
