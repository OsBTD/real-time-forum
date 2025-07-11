import * as auth from './auth.js';
import * as posts from './posts.js';
import * as chat from './chat.js';
import { showError } from './utils.js';

let currentUser = null;

export async function initApp() {
    try {
        currentUser = await auth.checkSession();
        if (currentUser) {
            renderMainApp();
        } else {
            renderAuthPage();
        }
    } catch (error) {
        showError('Failed to initialize app');
        renderAuthPage();
    }
}

function renderAuthPage() {
    document.getElementById('app').innerHTML = `
        <div class="auth-container">
            <div class="tabs">
                <button id="login-tab" class="active">Login</button>
                <button id="register-tab">Register</button>
            </div>
            <div id="login-form" class="form active">
                <input type="text" id="login-email" placeholder="Email or Nickname" required>
                <input type="password" id="login-password" placeholder="Password" required>
                <button id="login-btn">Login</button>
                <div id="login-error" class="error"></div>
            </div>
            <div id="register-form" class="form">
                <input type="text" id="reg-firstname" placeholder="First Name" required>
                <input type="text" id="reg-lastname" placeholder="Last Name" required>
                <input type="email" id="reg-email" placeholder="Email" required>
                <input type="text" id="reg-nickname" placeholder="Nickname" required>
                <input type="password" id="reg-password" placeholder="Password" required>
                <input type="number" id="reg-age" placeholder="Age" min="13" required>
                <select id="reg-gender" required>
                    <option value="">Select Gender</option>
                    <option value="male">Male</option>
                    <option value="female">Female</option>
                    <option value="other">Other</option>
                </select>
                <button id="register-btn">Register</button>
                <div id="register-error" class="error"></div>
            </div>
        </div>
    `;

    // Add event listeners
    document.getElementById('login-tab').addEventListener('click', () => switchTab('login'));
    document.getElementById('register-tab').addEventListener('click', () => switchTab('register'));
    document.getElementById('login-btn').addEventListener('click', handleLogin);
    document.getElementById('register-btn').addEventListener('click', handleRegister);
}

function switchTab(tab) {
    document.querySelectorAll('.form').forEach(f => f.classList.remove('active'));
    document.querySelectorAll('.tabs button').forEach(b => b.classList.remove('active'));

    if (tab === 'login') {
        document.getElementById('login-form').classList.add('active');
        document.getElementById('login-tab').classList.add('active');
    } else {
        document.getElementById('register-form').classList.add('active');
        document.getElementById('register-tab').classList.add('active');
    }
}

async function handleLogin() {
    const emailOrNickname = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        await auth.login(emailOrNickname, password);
        initApp(); // Reload app
    } catch (error) {
        showError('login-error', error.message);
    }
}

async function handleRegister() {
    const user = {
        firstName: document.getElementById('reg-firstname').value,
        lastName: document.getElementById('reg-lastname').value,
        email: document.getElementById('reg-email').value,
        nickname: document.getElementById('reg-nickname').value,
        password: document.getElementById('reg-password').value,
        age: parseInt(document.getElementById('reg-age').value),
        gender: document.getElementById('reg-gender').value
    };

    try {
        await auth.register(user);
        showError('register-error', 'Registration successful. Please login.', 'success');
        switchTab('login');
    } catch (error) {
        showError('register-error', error.message);
    }
}

function renderMainApp() {
    document.getElementById('app').innerHTML = `
        <header>
            <h1>Real-Time Forum</h1>
            <div class="user-info">
                <span>Welcome, ${currentUser.nickname}</span>
                <button id="logout-btn">Logout</button>
            </div>
        </header>
        <main>
            <div class="sidebar">
                <div class="online-users">
                    <h2>Online Users</h2>
                    <ul id="online-users-list"></ul>
                </div>
                <div class="categories">
                    <h2>Categories</h2>
                    <ul>
                        <li>General</li>
                        <li>Technology</li>
                        <li>Gaming</li>
                        <!-- More categories -->
                    </ul>
                </div>
            </div>
            <div class="content">
                <div class="posts-container">
                    <h2>Recent Posts</h2>
                    <div id="posts-list"></div>
                    <button id="load-more">Load More</button>
                </div>
                <div class="chat-container hidden">
                    <div class="chat-header">
                        <h2 id="chat-title">Chat</h2>
                        <button id="close-chat">✕</button>
                    </div>
                    <div id="chat-messages"></div>
                    <div class="chat-input">
                        <input type="text" id="message-input" placeholder="Type a message...">
                        <button id="send-message">Send</button>
                    </div>
                </div>
            </div>
        </main>
    `;

    // Load initial data
    posts.loadPosts();
    chat.initChat();

    // Event listeners
    document.getElementById('logout-btn').addEventListener('click', auth.logout);
    document.getElementById('load-more').addEventListener('click', posts.loadMorePosts);
}