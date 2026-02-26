// session.ts
// Utility to abstract whether tokens are stored permanently (localStorage) or temporarily (sessionStorage).

export const REMEMBER_ME_KEY = 'settings_remember_me';

/**
 * Gets the current store based on the "Remember Me" preference.
 * Defaults to localStorage if the preference isn't explicitly false.
 */
export const getStorage = (): Storage | null => {
    if (typeof window === 'undefined') return null;
    // Read preference from localStorage so it persists across sessions.
    // We want to default to true (Automatic Sign-in on) to maintain original behavior unless explicitly turned off.
    const rememberPrefs = localStorage.getItem(REMEMBER_ME_KEY);
    const shouldRemember = rememberPrefs === null || rememberPrefs !== 'false';

    return shouldRemember ? window.localStorage : window.sessionStorage;
};

export const setSessionItem = (key: string, value: string): void => {
    const store = getStorage();
    if (store) store.setItem(key, value);
};

export const getSessionItem = (key: string): string | null => {
    // Always check both stores to be safe, starting with the preferred one.
    if (typeof window === 'undefined') return null;
    const store = getStorage();

    if (store) {
        const val = store.getItem(key);
        if (val) return val;
    }

    // Fallback: check the other store if not found in the preferred one
    // This prevents losing a session if the user toggles the setting mid-session
    const otherStore = store === window.localStorage ? window.sessionStorage : window.localStorage;
    if (otherStore) return otherStore.getItem(key);

    return null;
};

export const removeSessionItem = (key: string): void => {
    if (typeof window === 'undefined') return;
    window.localStorage.removeItem(key);
    window.sessionStorage.removeItem(key);
};

export const clearSession = (): void => {
    removeSessionItem('access_token');
    removeSessionItem('refresh_token');
    removeSessionItem('user');
};
