const payButton = document.getElementById("pay-button");
const amountInput = document.getElementById("amount");
const sumupCard = document.getElementById("sumup-card");
const messageDiv = document.getElementById("message");
const loadingDiv = document.getElementById("loading");

// Display a message to the user
function showMessage(text, type) {
	messageDiv.textContent = text;
	messageDiv.className = `message ${type}`;
	messageDiv.style.display = "block";
}

// Handle payment button click
payButton.addEventListener("click", async () => {
	const amount = parseFloat(amountInput.value);

	if (!amount || amount <= 0) {
		showMessage("Please enter a valid amount", "error");
		return;
	}

	// Disable form and show loading
	payButton.disabled = true;
	messageDiv.style.display = "none";
	sumupCard.style.display = "none";
	loadingDiv.style.display = "block";

	try {
		// Create checkout on the server
		const response = await fetch("/checkout", {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({ amount: amount }),
		});

		if (!response.ok) {
			throw new Error("Failed to create checkout");
		}

		const data = await response.json();

		// Hide loading and show card widget
		loadingDiv.style.display = "none";
		sumupCard.style.display = "block";

		// Mount SumUp card widget
		SumUpCard.mount({
			id: "sumup-card",
			checkoutId: data.checkoutId,
			onResponse: (type, body) => {
				console.log("Payment response type:", type);
				console.log("Payment response body:", body);

				// Handle payment response
				if (type === "success") {
					showMessage(
						"Payment successful! Transaction ID: " +
							(body.transaction_id || "N/A"),
						"success",
					);
					sumupCard.style.display = "none";

					// Reset form after 3 seconds
					setTimeout(() => {
						amountInput.value = "10.00";
						payButton.disabled = false;
						messageDiv.style.display = "none";
					}, 3000);
				} else if (type === "error") {
					showMessage(
						`Payment failed: ${body.message || "Unknown error"}`,
						"error",
					);
					sumupCard.style.display = "none";
					payButton.disabled = false;
				}
			},
		});
	} catch (error) {
		loadingDiv.style.display = "none";
		showMessage(`Error: ${error.message}`, "error");
		payButton.disabled = false;
	}
});
