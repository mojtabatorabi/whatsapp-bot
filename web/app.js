const chat = document.getElementById("chat");

const messageInput = document.getElementById("message");

const sendButton = document.getElementById("send");

const typing = document.getElementById("typing");


// =========================
// User ID
// =========================

let userId = localStorage.getItem("chat_user_id");

if (!userId) {

    userId =
        "web-" +
        crypto.randomUUID();

    localStorage.setItem(
        "chat_user_id",
        userId
    );
}


// =========================
// Add Message
// =========================

function addMessage(
    message,
    type
) {

    const wrapper =
        document.createElement("div");

    wrapper.className =
        "message " + type;


    const bubble =
        document.createElement("div");

    bubble.className =
        "bubble";

    bubble.textContent =
        message;


    wrapper.appendChild(
        bubble
    );


    chat.appendChild(
        wrapper
    );


    scrollToBottom();
}


// =========================
// Scroll
// =========================

function scrollToBottom() {

    chat.scrollTop =
        chat.scrollHeight;
}


// =========================
// Typing
// =========================

function showTyping() {

    typing.classList.remove(
        "hidden"
    );

    scrollToBottom();
}


function hideTyping() {

    typing.classList.add(
        "hidden"
    );
}


// =========================
// Send Message
// =========================

async function sendMessage() {

    const message =
        messageInput.value.trim();


    if (!message) {
        return;
    }


    // Disable UI

    messageInput.disabled = true;

    sendButton.disabled = true;


    // Show user message

    addMessage(
        message,
        "user"
    );


    // Clear input

    messageInput.value = "";


    // Show typing

    showTyping();


    try {

        const response =
            await fetch(
                "/api/chat",
                {
                    method: "POST",

                    headers: {
                        "Content-Type":
                            "application/json"
                    },

                    body: JSON.stringify({
                        user: userId,
                        message: message
                    })
                }
            );


        if (!response.ok) {

            const errorText =
                await response.text();

            throw new Error(
                errorText
            );
        }


        const data =
            await response.json();


        hideTyping();


        addMessage(
            data.message,
            "assistant"
        );


    } catch (error) {

        hideTyping();


        addMessage(
            "❌ خطا در ارتباط با سرور: " +
            error.message,
            "assistant"
        );

        console.error(
            error
        );

    } finally {

        messageInput.disabled =
            false;

        sendButton.disabled =
            false;

        messageInput.focus();
    }
}


// =========================
// Enter
// =========================

messageInput.addEventListener(
    "keydown",
    function (event) {

        if (
            event.key === "Enter" &&
            !event.shiftKey
        ) {

            event.preventDefault();

            sendMessage();
        }
    }
);


// =========================
// Auto Resize
// =========================

messageInput.addEventListener(
    "input",
    function () {

        this.style.height =
            "auto";

        this.style.height =
            Math.min(
                this.scrollHeight,
                120
            ) + "px";
    }
);
