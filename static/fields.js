let additionalFields = document.querySelector("#additional-fields");
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

	fieldName.addEventListener("change", function(ev) {
		fieldValue.name = fieldName.value;
	});

	additionalFields.appendChild(fieldContainer);

	fieldName.focus();

	// prevent form submit
	ev.preventDefault();
});
