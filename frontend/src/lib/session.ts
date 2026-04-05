// session.ts
// Utility to abstract whether tokens are stored permanently (localStorage) or temporarily (sessionStorage).

export const getRememberMeKey = (userId: string) => `settings_remember_me_${userId}`;

/**
 * Gets the current store based on the active user's "Remember Me" preference.
 * Defaults to localStorage if the preference isn't explicitly false.
 */
export const getStorage = (): Storage | null => {
    if (typeof window === 'undefined') return null;
    
    // First, if there's an active session in sessionStorage, keep using it
    if (window.sessionStorage.getItem('access_token')) {
        return window.sessionStorage;
    }
    
    // Then check if we had a stored user's preference
    const u = window.sessionStorage.getItem('user') || window.localStorage.getItem('user');
    if (u) {
        try {
            const user = JSON.parse(u);
            const pref = window.localStorage.getItem(getRememberMeKey(user.id));
            if (pref === 'false') {
                return window.sessionStorage;
            }
        } catch(e) {}
    }

    return window.localStorage;
};

export const setSessionItem = (key: string, value: string): void => {
    const store = getStorage();
    if (store) store.setItem(key, value);
};

export const getSessionItem = (key: string): string | null => {
    if (typeof window === 'undefined') return null;
    return window.sessionStorage.getItem(key) || window.localStorage.getItem(key) || null;
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
