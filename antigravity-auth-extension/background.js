// Background Service Worker
// Listens for OAuth callback URL and extracts auth code automatically

const CALLBACK_URL_PATTERN = 'http://localhost:51121/oauth-callback';

// Listen for tab updates
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
    // Check if URL matches our callback pattern
    if (changeInfo.url && changeInfo.url.startsWith(CALLBACK_URL_PATTERN)) {
        console.log('Detected OAuth callback URL:', changeInfo.url);

        // Extract auth code from URL
        try {
            const url = new URL(changeInfo.url);
            const code = url.searchParams.get('code');
            const state = url.searchParams.get('state');
            const error = url.searchParams.get('error');

            if (error) {
                // Store error
                chrome.storage.local.set({
                    authCallback: { error: error, timestamp: Date.now() }
                });
            } else if (code) {
                // Store the code for popup to process
                chrome.storage.local.set({
                    authCallback: { code: code, state: state, url: changeInfo.url, timestamp: Date.now() }
                });

                // Show notification badge
                chrome.action.setBadgeText({ text: '!' });
                chrome.action.setBadgeBackgroundColor({ color: '#38ef7d' });
            }

            // Close the error tab after a short delay
            setTimeout(() => {
                chrome.tabs.remove(tabId).catch(() => { });
            }, 500);

        } catch (e) {
            console.error('Error parsing callback URL:', e);
        }
    }
});

// Clear badge when popup is opened
chrome.action.onClicked.addListener(() => {
    chrome.action.setBadgeText({ text: '' });
});
