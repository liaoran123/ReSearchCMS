// Custom JS
document.addEventListener('DOMContentLoaded', function() {
    // Add any custom JavaScript here
    console.log('Search Engine Web App Loaded');
});

// Function to change language
function changeLanguage(lang) {
    const currentUrl = window.location.href;
    const newUrl = updateQueryStringParameter(currentUrl, 'lang', lang);
    window.location.href = newUrl;
}

// Helper function to update query string parameters
function updateQueryStringParameter(uri, key, value) {
    const re = new RegExp("([?&])" + key + "=.*?(&|$)", "i");
    const separator = uri.indexOf('?') !== -1 ? "&" : "?";
    if (uri.match(re)) {
        return uri.replace(re, '$1' + key + "=" + value + '$2');
    } else {
        return uri + separator + key + "=" + value;
    }
}