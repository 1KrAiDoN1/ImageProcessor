/**
 * Frontend Configuration
 * 
 * Измените эти настройки в соответствии с вашей средой
 */

const CONFIG = {
    // API Configuration
    api: {
        // В production используем относительный путь (через nginx proxy)
        // В development - прямое подключение к backend
        baseURL: window.location.hostname === 'localhost' 
            ? 'http://localhost:8080/api/v1' 
            : '/api/v1',
        timeout: 30000, // 30 seconds
        
        // Retry configuration
        retry: {
            maxAttempts: 3,
            delay: 1000 // 1 second
        },
        
        // Polling configuration
        polling: {
            interval: 2000, // 2 seconds
            maxAttempts: 60 // 2 minutes total
        }
    },
    
    // Upload Configuration
    upload: {
        maxFileSize: 32 * 1024 * 1024, // 32 MB
        allowedTypes: ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'],
        allowedExtensions: ['.jpg', '.jpeg', '.png', '.gif', '.webp']
    },
    
    // Gallery Configuration
    gallery: {
        pageSize: 12,
        loadMoreOnScroll: true,
        imagePlaceholder: 'data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' width=\'250\' height=\'250\'%3E%3Crect fill=\'%23ddd\' width=\'250\' height=\'250\'/%3E%3Ctext fill=\'%23999\' x=\'50%25\' y=\'50%25\' text-anchor=\'middle\' dy=\'.3em\'%3EЗагрузка...%3C/text%3E%3C/svg%3E'
    },
    
    // UI Configuration
    ui: {
        toastDuration: 5000, // 5 seconds
        animationDuration: 300, // 300ms
        theme: 'light', // 'light' or 'dark'
        language: 'ru' // 'ru' or 'en'
    },
    
    // Statistics Configuration
    statistics: {
        autoRefresh: true,
        refreshInterval: 30000 // 30 seconds
    },
    
    // Debug Configuration
    debug: {
        enabled: false,
        logLevel: 'info' // 'debug', 'info', 'warn', 'error'
    }
};

// Freeze configuration to prevent modifications
Object.freeze(CONFIG);

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = CONFIG;
}

