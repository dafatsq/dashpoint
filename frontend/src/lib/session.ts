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
    if (typeof window === 'undefined') return null;
    const store = getStorage();
    if (store) return store.getItem(key);
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

/**
 * Migrate session tokens from one storage to the other.
 * Call this when the user changes the "Remember Me" preference.
 * @param toLocalStorage - if true, migrate from sessionStorage to localStorage; if false, the reverse.
 */
export const migrateSession = (toLocalStorage: boolean): void => {
    if (typeof window === 'undefined') return;

    const source = toLocalStorage ? window.sessionStorage : window.localStorage;
    const target = toLocalStorage ? window.localStorage : window.sessionStorage;
    const keys = ['access_token', 'refresh_token', 'user'];

    for (const key of keys) {
        const value = source.getItem(key);
        if (value) {
            target.setItem(key, value);
            source.removeItem(key);
        }
    }
};
