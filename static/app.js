let socket

const connectWebSocket = () => {
    socket = new WebSocket(`ws://${location.host}/ws`)

    socket.onopen = () => console.log("✅ WebSocket connected!");
    socket.onerror = (e) => console.error("❌ WebSocket error:", e);
    socket.onclose = (e) => {
        console.warn("⚠️ WebSocket closed!", e);
        setTimeout(connectWebSocket, 1000);
    }

    socket.onmessage = e => {
        console.log("📩 Message from server:", e.data); // 👀
        const chatBox = this.document.getElementById("chat-box")
        const line = document.createElement('div')
        line.classList.add('line')

        if (e.data.includes('has joined')) {
            line.textContent = e.data
            line.style.fontStyle = 'italic'
            line.style.color = 'green'
        } else if (e.data.includes('has left')) {
            line.textContent = e.data
            line.style.fontStyle = 'italic'
            line.style.color = 'red'
        }
        else {
            line.textContent = e.data
        }
        chatBox.appendChild(line)
        chatBox.scrollTop = chatBox.scrollHeight
    }
}

connectWebSocket()

async function fetchMessages() {

    const res = await fetch("/messages")
    const messages = await res.json()

    const chatBox = this.document.getElementById("chat-box")

    messages && messages.length ?
        messages.forEach(msg => {
            const line = document.createElement("div")
            line.classList.add('line')
            line.textContent = `[${msg.created_at}] ${msg.username} : ${msg.message}`
            chatBox.appendChild(line)
        }) : null

}


function sendMessage() {

    const input = document.getElementById('msg-input')
    const msg = input.value.trim()
    console.log(socket.readyState);

    if (!msg) return

    // await fetch('/send', {
    //     method: 'POST',
    //     headers: {
    //         'Content-Type': 'application/x-www-form-urlencoded'
    //     },
    //     body: `message=${encodeURIComponent(msg)}`
    // })
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(msg);
        input.value = "";
    } else {
        alert("WebSocket is not connected!");
    }
}
