export const showError = (elem, msg) => {
    elem.hidden = false;
    elem.classList.add("error");
    elem.innerHTML = msg;
};
export const showResult = (elem, msg) => {
    elem.hidden = false;
    elem.classList.remove("error");
    elem.innerHTML = msg;
};
export const hide = (elem) => {
    elem.hidden = true;
    elem.classList.remove("error")
}
