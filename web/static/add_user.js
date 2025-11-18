import { showError, showResult, hide } from "./status.js"

const inputUsername = document.getElementById("input-username");
const form = document.getElementById("form");
const status = document.getElementById("status")

form.addEventListener("submit", async (event) => {
    event.preventDefault();

    hide(status)

    const username = inputUsername.value;

    try {
        const resp = await fetch(
            "/api/groups/" + groupID + "/users",
            {
                method: "POST",
                header: {"Content-Type": "application/json"},
                body: JSON.stringify({"username": username}),
                credentials: "same-origin"
            }
        );

        if (!resp.ok) {
            const msg = await resp.text();
            showError(status, msg);
            console.log(`${resp.status}: ${msg}`);
        } else {
            window.location.href = `/groups/${groupID}`;
        }
    } catch (e) {
        console.log(e)
    }
});

