const inputUsername = document.getElementById("input-username");
const inputPassword = document.getElementById("input-password");
const form = document.getElementById("form");

form.addEventListener("submit", async (event) => {
    event.preventDefault();

    const username = inputUsername.value;
    const password = inputPassword.value;

    const resp = await fetch(
        "/api/login",
        {
            method: "POST",
            header: {"Content-Type": "application/json"},
            body: JSON.stringify({"username": username, "password": password}),
            credentials: "same-origin"
        }
    );

    if (!resp.ok) {
        console.log(resp.status)
    } else {
        window.location.href = "/dashboard"
    }
});
