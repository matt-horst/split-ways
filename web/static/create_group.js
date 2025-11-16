import { showError, showResult, hide } from "./status.js"

const inputName = document.getElementById("input-name");
const form = document.getElementById("form");
const status = document.getElementById("status")

form.addEventListener("submit", async (event) => {
    event.preventDefault();

    hide(status)

    const name = inputName.value;

    try {
        const resp = await fetch(
            "/api/groups",
            {
                method: "POST",
                header: {"Content-Type": "application/json"},
                body: JSON.stringify({"name": name}),
                credentials: "same-origin"
            }
        );

        if (!resp.ok) {
            const msg = await resp.text();
            showError(status, msg);
            console.log(`${resp.status}: ${msg}`);
        } else {
            const body = await resp.json()
            console.log(JSON.stringify(body))
            window.location.href = "/dashboard";
        }
    } catch (e) {
        console.log(e)
    }
});

