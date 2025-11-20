import { showError, showResult, hide } from "./status.js"

const inputUsername = document.getElementById("input-username");
const inputNewName = document.getElementById("input-new-name");
const addUserForm = document.getElementById("add-user-form");
const renameGroupForm = document.getElementById("rename-group-form");
const status = document.getElementById("status")
const deleteButtons = document.querySelectorAll(".btn-delete");
const deleteGroupForm = document.getElementById("delete-group-form");

addUserForm.addEventListener("submit", async (event) => {
    event.preventDefault();

    hide(status)

    const username = inputUsername.value.trim();

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


deleteButtons.forEach(btn => {
    const userID = btn.dataset.id;

    btn.addEventListener("click", async (event) => {
        try {
            const resp = await fetch(
                `/api/groups/${groupID}/users`,
                {
                    method: "DELETE",
                    header: {"Content-Type": "application/json"},
                    body: JSON.stringify({"id": userID}),
                    credentials: "same-origin"
                }
            );

            if (resp.ok) {
                window.location.href = `/groups/${groupID}`
            } else {
                console.log(await resp.text());
            }
        } catch (e) {
            console.log(e);
        }
    });
});

renameGroupForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const name = inputNewName.value.trim();

    try {
        const resp = await fetch(
            "/api/groups/" + groupID,
            {
                method: "PUT",
                header: {"Content-Type": "application/json"},
                body: JSON.stringify({"name": name}),
                credentials: "same-origin"
            }
        );

        if (!resp.ok) {
            const msg = await resp.text();
            console.log(`${resp.status}: ${msg}`);
        } else {
            window.location.href = `/groups/${groupID}`;
        }
    } catch (e) {
        console.log(e)
    }
});

deleteGroupForm.addEventListener("submit", async (event) => {
    event.preventDefault();

    try {
        const resp = await fetch(
            "/api/groups/" + groupID,
            {
                method: "DELETE",
                credentials: "same-origin"
            }
        );

        if (resp.ok) {
            window.location.href = "/";
        } else {
            const msg = await resp.text();
            console.log(`${resp.status}: ${msg}`);
        }
    } catch (e) {
        console.log(e);
    }
});
