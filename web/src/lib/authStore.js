import { writable } from 'svelte/store';

// We try to recover the token and user from localStorage on startup
const storedToken = localStorage.getItem('token');
const storedUser = JSON.parse(localStorage.getItem('user') || 'null');

export const token = writable(storedToken);
export const user = writable(storedUser);

// Subscriptions to keep localStorage synchronized
token.subscribe(value => {
    if (value) localStorage.setItem('token', value);
    else localStorage.removeItem('token');
});

user.subscribe(value => {
    if (value) localStorage.setItem('user', JSON.stringify(value));
    else localStorage.removeItem('user');
});

export const logout = () => {
    token.set(null);
    user.set(null);
};