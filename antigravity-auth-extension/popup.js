// Antigravity Auth Chrome Extension - Auto-detect Version
// Automatically detects OAuth callback URL

const CONFIG = {
    clientId: '1071006060591-tmhssin2h21lcre235vtolojh4g403ep.apps.googleusercontent.com',
    clientSecret: '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',
    redirectUri: 'http://localhost:51121/oauth-callback',
    scopes: [
        'https://www.googleapis.com/auth/cloud-platform',
        'https://www.googleapis.com/auth/userinfo.email',
        'https://www.googleapis.com/auth/userinfo.profile',
        'https://www.googleapis.com/auth/cclog',
        'https://www.googleapis.com/auth/experimentsandconfigs'
    ],
    tokenEndpoint: 'https://oauth2.googleapis.com/token',
    userInfoEndpoint: 'https://www.googleapis.com/oauth2/v1/userinfo?alt=json'
};

// State
let authData = null;

// DOM Elements
const loginSection = document.getElementById('login-section');
const loadingSection = document.getElementById('loading-section');
const successSection = document.getElementById('success-section');
const errorSection = document.getElementById('error-section');

const loginBtn = document.getElementById('login-btn');
const downloadBtn = document.getElementById('download-btn');
const logoutBtn = document.getElementById('logout-btn');
const retryBtn = document.getElementById('retry-btn');

const userEmailEl = document.getElementById('user-email');
const errorMessageEl = document.getElementById('error-message');
const statusText = document.getElementById('status-text');

// Event Listeners
loginBtn.addEventListener('click', startAuth);
downloadBtn.addEventListener('click', downloadAuthJson);
logoutBtn.addEventListener('click', resetState);
retryBtn.addEventListener('click', () => { resetState(); showSection(loginSection); });

// Check for pending auth callback on popup open
document.addEventListener('DOMContentLoaded', checkPendingCallback);

// Show/Hide UI Sections
function showSection(section) {
    [loginSection, loadingSection, successSection, errorSection].forEach(s => s.classList.add('hidden'));
    section.classList.remove('hidden');
}

// Check if there's a pending auth callback
async function checkPendingCallback() {
    // Clear badge
    chrome.action.setBadgeText({ text: '' });

    const result = await chrome.storage.local.get(['authCallback']);
    const callback = result.authCallback;

    if (callback && callback.code) {
        // Check if callback is recent (within 5 minutes)
        if (Date.now() - callback.timestamp < 5 * 60 * 1000) {
            // Process the callback
            await processAuthCode(callback.code, callback.state);
            // Clear the stored callback
            chrome.storage.local.remove('authCallback');
            return;
        } else {
            // Expired, clear it
            chrome.storage.local.remove('authCallback');
        }
    }

    // No pending callback, show login
    showSection(loginSection);
}

// Generate random state
function generateState() {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return Array.from(array, b => b.toString(16).padStart(2, '0')).join('');
}

// Start OAuth Flow
function startAuth() {
    const state = generateState();
    chrome.storage.local.set({ oauth_state: state });

    // Build auth URL
    const authUrl = new URL('https://accounts.google.com/o/oauth2/v2/auth');
    authUrl.searchParams.set('client_id', CONFIG.clientId);
    authUrl.searchParams.set('redirect_uri', CONFIG.redirectUri);
    authUrl.searchParams.set('response_type', 'code');
    authUrl.searchParams.set('scope', CONFIG.scopes.join(' '));
    authUrl.searchParams.set('access_type', 'offline');
    authUrl.searchParams.set('prompt', 'consent select_account');
    authUrl.searchParams.set('state', state);

    // Open in new tab
    chrome.tabs.create({ url: authUrl.toString() });

    // Show loading with instruction
    showSection(loadingSection);
    if (statusText) {
        statusText.innerHTML = 'Complete login in the opened tab...<br><small>Extension will auto-detect when done</small>';
    }

    // Start polling for callback
    pollForCallback();
}

// Poll for auth callback
async function pollForCallback() {
    const maxAttempts = 120; // 2 minutes
    const interval = 1000; // 1 second

    for (let i = 0; i < maxAttempts; i++) {
        const result = await chrome.storage.local.get(['authCallback']);
        const callback = result.authCallback;

        if (callback) {
            if (callback.error) {
                chrome.storage.local.remove('authCallback');
                errorMessageEl.textContent = `OAuth error: ${callback.error}`;
                showSection(errorSection);
                return;
            }

            if (callback.code) {
                await processAuthCode(callback.code, callback.state);
                chrome.storage.local.remove('authCallback');
                return;
            }
        }

        // Update status
        if (statusText && i % 5 === 0) {
            statusText.innerHTML = `Waiting for login... (${Math.floor((maxAttempts - i) / 60)}:${String((maxAttempts - i) % 60).padStart(2, '0')})<br><small>Extension will auto-detect when done</small>`;
        }

        await new Promise(r => setTimeout(r, interval));
    }

    // Timeout
    errorMessageEl.textContent = 'Authentication timed out. Please try again.';
    showSection(errorSection);
}

// Process auth code
async function processAuthCode(code, state) {
    showSection(loadingSection);
    if (statusText) statusText.textContent = 'Exchanging tokens...';

    try {
        // Verify state
        const stored = await chrome.storage.local.get(['oauth_state']);
        if (stored.oauth_state && state && stored.oauth_state !== state) {
            throw new Error('State mismatch - security check failed');
        }

        // Exchange code for tokens
        const tokenData = await exchangeCodeForTokens(code);

        if (statusText) statusText.textContent = 'Fetching user info...';

        // Fetch user info
        const userInfo = await fetchUserInfo(tokenData.access_token);

        // Build auth data
        const now = Date.now();
        const expiresIn = tokenData.expires_in || 3599;
        const expiredDate = new Date(now + expiresIn * 1000);

        authData = {
            access_token: tokenData.access_token,
            email: userInfo.email || '',
            expired: formatDateWithTimezone(expiredDate),
            expires_in: expiresIn,
            refresh_token: tokenData.refresh_token || '',
            timestamp: now,
            type: 'antigravity'
        };

        // Show success
        userEmailEl.textContent = authData.email || 'Unknown user';
        showSection(successSection);
        displayTokenInfo();

        // Clear stored state
        chrome.storage.local.remove('oauth_state');

    } catch (error) {
        console.error('Auth error:', error);
        errorMessageEl.textContent = error.message || 'Unknown error occurred';
        showSection(errorSection);
    }
}

// Exchange auth code for tokens
async function exchangeCodeForTokens(code) {
    const response = await fetch(CONFIG.tokenEndpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
            code: code,
            client_id: CONFIG.clientId,
            client_secret: CONFIG.clientSecret,
            redirect_uri: CONFIG.redirectUri,
            grant_type: 'authorization_code'
        })
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error_description || `Token exchange failed: ${response.status}`);
    }

    return response.json();
}

// Fetch user info
async function fetchUserInfo(accessToken) {
    const response = await fetch(CONFIG.userInfoEndpoint, {
        headers: { 'Authorization': `Bearer ${accessToken}` }
    });

    if (!response.ok) {
        throw new Error(`Failed to fetch user info: ${response.status}`);
    }

    return response.json();
}

// Format date with timezone
function formatDateWithTimezone(date) {
    const offset = -date.getTimezoneOffset();
    const sign = offset >= 0 ? '+' : '-';
    const hours = String(Math.floor(Math.abs(offset) / 60)).padStart(2, '0');
    const minutes = String(Math.abs(offset) % 60).padStart(2, '0');

    const pad = n => String(n).padStart(2, '0');

    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}${sign}${hours}:${minutes}`;
}

// Display token info
function displayTokenInfo() {
    const tokenInfoDiv = document.getElementById('token-info');
    if (tokenInfoDiv && authData) {
        tokenInfoDiv.innerHTML = `
      <div class="token-display">
        <div class="token-row">
          <span class="token-label">Email:</span>
          <span class="token-value">${authData.email || 'N/A'}</span>
        </div>
        <div class="token-row">
          <span class="token-label">Expires:</span>
          <span class="token-value">${authData.expired || 'N/A'}</span>
        </div>
        <div class="token-row">
          <span class="token-label">Access Token:</span>
          <span class="token-value token-truncate">${(authData.access_token || '').substring(0, 25)}...</span>
        </div>
        <div class="token-row">
          <span class="token-label">Refresh Token:</span>
          <span class="token-value token-truncate">${authData.refresh_token ? authData.refresh_token.substring(0, 25) + '...' : 'N/A'}</span>
        </div>
      </div>
    `;
        tokenInfoDiv.classList.remove('hidden');
    }
}

// Download auth JSON
function downloadAuthJson() {
    if (!authData) {
        alert('No auth data available');
        return;
    }

    const email = authData.email || 'unknown';
    const sanitizedEmail = email.replace(/@/g, '_').replace(/\./g, '_');
    const filename = `antigravity-${sanitizedEmail}.json`;

    const jsonContent = JSON.stringify(authData);
    const blob = new Blob([jsonContent], { type: 'application/json' });
    const url = URL.createObjectURL(blob);

    chrome.downloads.download({
        url: url,
        filename: filename,
        saveAs: true
    }, () => {
        if (chrome.runtime.lastError) {
            alert('Download failed: ' + chrome.runtime.lastError.message);
        }
        URL.revokeObjectURL(url);
    });
}

// Reset state
function resetState() {
    authData = null;
    chrome.storage.local.remove(['oauth_state', 'authCallback']);
    chrome.action.setBadgeText({ text: '' });
    const tokenInfoDiv = document.getElementById('token-info');
    if (tokenInfoDiv) tokenInfoDiv.classList.add('hidden');
    showSection(loginSection);
}
