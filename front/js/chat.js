let socket = null;
let currentChatUser = null;

export function initChat() {
    // Connect to WebSocket
    socket = new WebSocket(`ws://${window.location.host}/ws`);

    socket.onopen = () => {
        console.log('WebSocket connected');
        // Request online users
        socket.send(JSON.stringify({ type: 'get_online_users' }));
    };

    socket.onmessage = (event) => {
        const message = JSON.parse(event.data);
        handleSocketMessage(message);
    };

    socket.onclose = () => {
        console.log('WebSocket disconnected');
    };

    // Event listeners for UI
    document.querySelectorAll('.user-item').forEach(user => {
        user.addEventListener('click', () => openChat(user.dataset.userId));
    });

    document.getElementById('send-message').addEventListener('click', sendMessage);
    document.getElementById('message-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') sendMessage();
    });
}

function openChat(userId) {
    currentChatUser = userId;
    document.querySelector('.chat-container').classList.remove('hidden');
    document.getElementById('chat-title').textContent = `Chat with ${userId}`;
    loadChatHistory(userId);
}

function loadChatHistory(userId) {
    fetch(`/api/messages?userId=${userId}`)
        .then(response => response.json())
        .then(messages => {
            const chatContainer = document.getElementById('chat-messages');
            chatContainer.innerHTML = '';

            messages.forEach(msg => {
                const messageElement = document.createElement('div');
                messageElement.classList.add('message', msg.sender === currentUser.id ? 'sent' : 'received');
                messageElement.innerHTML = `
                    <div class="message-content">${msg.content}</div>
                    <div class="message-time">${new Date(msg.timestamp).toLocaleTimeString()}</div>
                `;
                chatContainer.appendChild(messageElement);
            });

            chatContainer.scrollTop = chatContainer.scrollHeight;
        });
}

function sendMessage() {
    const input = document.getElementById('message-input');
    const content = input.value.trim();

    if (content && currentChatUser) {
        const message = {
            receiver: currentChatUser,
            content: content,
            timestamp: new Date().toISOString()
        };

        // Send via WebSocket for real-time
        socket.send(JSON.stringify({
            type: 'chat_message',
            payload: message
        }));

        // Also save to database
        fetch('/api/messages', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(message)
        });

        input.value = '';
    }
}

function handleSocketMessage(message) {
    switch (message.type) {
        case 'online_users':
            updateOnlineUsers(message.payload);
            break;
        case 'chat_message':
            if (message.payload.sender === currentChatUser) {
                appendMessage(message.payload);
            }
            break;
        case 'user_typing':
            // Show typing indicator
            break;
    }
}

function updateOnlineUsers(users) {
    const list = document.getElementById('online-users-list');
    list.innerHTML = '';

    users.forEach(user => {
        const item = document.createElement('li');
        item.textContent = user.nickname;
        item.classList.add('user-item');
        item.dataset.userId = user.id;
        list.appendChild(item);
    });
}

function appendMessage(message) {
    const chatContainer = document.getElementById('chat-messages');
    const messageElement = document.createElement('div');
    messageElement.classList.add('message', 'received');
    messageElement.innerHTML = `
        <div class="message-content">${message.content}</div>
        <div class="message-time">${new Date(message.timestamp).toLocaleTimeString()}</div>
    `;
    chatContainer.appendChild(messageElement);
    chatContainer.scrollTop = chatContainer.scrollHeight;
}
