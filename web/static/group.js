const editButtons = document.querySelectorAll(".btn-edit");
const deleteButtons = document.querySelectorAll(".btn-delete");

editButtons.forEach(btn => {
    const txID = btn.dataset.id;

    btn.addEventListener("click", (event) => {
        window.location.href = `/edit?id=${txID}`
    });
});

deleteButtons.forEach(btn => {
    const txID = btn.dataset.id;

    btn.addEventListener("click", async (event) => {
        try {
            const resp = await fetch(
                `/api/groups/${groupID}/transactions`,
                {
                    method: "DELETE",
                    header: {"Content-Type": "application/json"},
                    body: JSON.stringify({"id": txID}),
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

