import { showError, showResult, hide } from "./status.js"

const inputPaidBy = document.getElementById("input-paid-by");
const inputPaidTo = document.getElementById("input-paid-to");
const inputAmount = document.getElementById("input-amount");
const form = document.getElementById("form");
const status = document.getElementById("status")

form.addEventListener("submit", async (event) => {
    event.preventDefault();

    hide(status)

    const paidBy = inputPaidBy.value.trim();
    const paidTo = inputPaidTo.value.trim();

    var amount = inputAmount.value.replace('$', '').trim();
    amount = amount ? amount : "-1";

    try {
        const resp = await fetch(
            "/api/groups/" + groupID + "/payments?id=" + transactionID,
            {
                method: "PUT",
                header: {"Content-Type": "application/json"},
                body: JSON.stringify(
                    {
                        "paid_by": paidBy,
                        "paid_to": paidTo,
                        "amount": amount
                    }
                ),
                credentials: "same-origin"
            }
        );

        if (!resp.ok) {
            const msg = await resp.text();
            showError(status, msg);
            console.log(`${resp.status}: ${msg}`);
        } else {
            window.location.href = "/groups/" + groupID;
        }
    } catch (e) {
        console.log(e)
    }
});

inputAmount.addEventListener("input", (e) => {
    let value = e.target.value;

    value = value.replace(/[^\d.]/g, '');

    e.target.value = value;
});

inputAmount.addEventListener("blur", (e) => {
    let value = e.target.value;

    // Remove illegal characters
    value = value.replace(/[^\d.]/g, '');

    // Remove leading zeros
    value = value.replace(/^0+/, '')

    if (value.includes('.')) {
        let parts = value.split('.');

        if (parts[1].length == 1) {
            parts[1] += "0";
        } else if (parts[1].length > 2) {
            parts[1] = parts[1][0] + parts[0][1];
        }

        value = `$${parts[0]}.${parts[1]}`;
    } else if (value) {
        value = `$${value}.00`
    } else {
        value = ""
    }

    e.target.value = value;
});
