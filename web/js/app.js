const msgBox = document.getElementById('messageBox');
const msgForm = document.getElementById('messageForm');
const msgInput = document.getElementById('msgInput');
const sendBtn = document.getElementById('sendBtn');

// 1. Foydalanuvchi identifikatsiyasi
const myUserId = prompt("User ID kiriting:", "1");

// 2. WebSocket ulanishi
const socket = new WebSocket(`ws://localhost:8080/api/v1/ws?user_id=${myUserId}`);

socket.onmessage = (event) => {
    try {
        const data = JSON.parse(event.data);
        if (data.sender_id == myUserId) return;
        appendMessage(data.content, 'other', data.sender_name || "Foydalanuvchi");
    } catch (err) { console.error("Xato:", err); }
};

// 3. Xabar yuborish
msgForm.onsubmit = (e) => {
    e.preventDefault();
    const text = msgInput.value.trim();
    if (!text) return;

    const payload = { to: "2", content: text }; // Test uchun
    socket.send(JSON.stringify(payload));

    appendMessage(text, 'my-message', 'Siz');
    msgInput.value = '';
    sendBtn.style.display = 'none';
};

// 4. Dinamik yuborish tugmasi
msgInput.addEventListener('input', () => {
    sendBtn.style.display = msgInput.value.trim() ? 'block' : 'none';
});

// 5. Xabarni chiqarish
function appendMessage(content, type, senderName) {
    const div = document.createElement('div');
    div.className = `message ${type}`;
    
    const time = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

    div.innerHTML = `
        <div style="font-size: 12px; font-weight: bold; margin-bottom: 4px; color: ${type === 'my-message' ? '#8bc34a' : '#4faaff'}">
            ${senderName}
        </div>
        <div class="text-content" style="word-break: break-word;">${content}</div>
        <div style="font-size: 10px; text-align: right; color: #708499; margin-top: 5px;">
            ${time} ${type === 'my-message' ? '<i class="fas fa-check-double"></i>' : ''}
        </div>
    `;
    
    msgBox.appendChild(div);
    msgBox.scrollTop = msgBox.scrollHeight;
}

// 6. Guruh yaratish (Sidebar'dagi tugma uchun)
async function createGroup() {
    const name = prompt("Guruh nomi:");
    if (!name) return;
    const res = await fetch('/api/v1/groups', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name })
    });
    if (res.ok) alert("Guruh ochildi!");
}