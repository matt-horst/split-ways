import { showError, showResult, hide } from "./status.js"
const status = document.getElementById("status")

try {
    const resp = await fetch(
        "/api/logout",
        {
            method: "POST",
            credentials: "same-origin"
        }
    );

    if (!resp.ok) {
        const msg = await resp.text();
        showError(status, msg);
        console.log(`${resp.status}: ${msg}`);
    } else {
        window.location.href = "/login"
    }
} catch (e) {
    console.log(e)
}

