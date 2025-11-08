import { showError, showResult, hide } from "./status.js"

const inputUsername = document.getElementById("input-username");
const inputPassword = document.getElementById("input-password");
const form = document.getElementById("form");
const status = document.getElementById("status")

form.addEventListener("submit", async (event) => {
    event.preventDefault();

    hide(status)

    const username = inputUsername.value;
    const password = inputPassword.value;

    try {
        const resp = await fetch(
            "/api/users",
            {
                method: "POST",
                header: {"Content-Type": "application/json"},
                body: JSON.stringify({"username": username, "password": password}),
                credentials: "same-origin"
            }
        );

        if (!resp.ok) {
            const msg = await resp.text();
            showError(status, msg);
            console.log(`${resp.status}: ${msg}`);
        } else {
            window.location.href = "/dashboard";
        }
    } catch (e) {
        console.log(e)
    }
});

