const API_BASE_URL = '/api';

export async function checkSession() {
    try {
        const response = await fetch(`${API_BASE_URL}/session`, {
            credentials: 'include'
        });

        if (response.ok) {
            return await response.json();
        }
        return null;
    } catch (error) {
        console.error('Session check failed:', error);
        return null;
    }
}

export async function login(emailOrNickname, password) {
    try {
        const response = await fetch(`${API_BASE_URL}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ emailOrNickname, password }),
            credentials: 'include'
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Login failed');
        }
    } catch (error) {
        throw error;
    }
}

export async function register(user) {
    try {
        const response = await fetch(`${API_BASE_URL}/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(user)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Registration failed');
        }
    } catch (error) {
        throw error;
    }
}

export async function logout() {
    try {
        await fetch(`${API_BASE_URL}/logout`, {
            method: 'POST',
            credentials: 'include'
        });
        window.location.reload();
    } catch (error) {
        console.error('Logout failed:', error);
    }
}