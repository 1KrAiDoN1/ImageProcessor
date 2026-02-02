/**
 * API Module - взаимодействие с backend
 */

class ImageProcessorAPI {
    constructor(config = CONFIG.api) {
        this.baseURL = config.baseURL;
        this.timeout = config.timeout;
        this.retryConfig = config.retry;
        this.pollingConfig = config.polling;
    }

    /**
     * Загрузить изображение с операциями
     */
    async uploadImage(file, operations) {
        const formData = new FormData();
        formData.append('image', file);
        
        if (operations && operations.length > 0) {
            formData.append('operations', JSON.stringify(operations));
        }

        try {
            const response = await fetch(`${this.baseURL}/images`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Upload failed');
            }

            return await response.json();
        } catch (error) {
            console.error('Upload error:', error);
            throw error;
        }
    }

    /**
     * Получить статус обработки изображения
     */
    async getImageStatus(imageId) {
        try {
            const response = await fetch(`${this.baseURL}/images/${imageId}/status`);
            
            if (!response.ok) {
                throw new Error('Failed to get image status');
            }

            return await response.json();
        } catch (error) {
            console.error('Get status error:', error);
            throw error;
        }
    }

    /**
     * Получить URL изображения
     */
    getImageURL(imageId, operation = 'original') {
        return `${this.baseURL}/images/${imageId}?operation=${operation}`;
    }

    /**
     * Получить presigned URL для изображения
     */
    async getPresignedURL(imageId, operation = 'original', expiry = 3600) {
        try {
            const response = await fetch(
                `${this.baseURL}/images/${imageId}/url?operation=${operation}&expiry=${expiry}`
            );
            
            if (!response.ok) {
                throw new Error('Failed to get presigned URL');
            }

            return await response.json();
        } catch (error) {
            console.error('Get presigned URL error:', error);
            throw error;
        }
    }

    /**
     * Удалить изображение
     */
    async deleteImage(imageId) {
        try {
            const response = await fetch(`${this.baseURL}/images/${imageId}`, {
                method: 'DELETE'
            });

            if (!response.ok) {
                throw new Error('Failed to delete image');
            }

            return await response.json();
        } catch (error) {
            console.error('Delete error:', error);
            throw error;
        }
    }

    /**
     * Получить список изображений
     */
    async listImages(limit = 20, offset = 0) {
        try {
            const response = await fetch(
                `${this.baseURL}/images?limit=${limit}&offset=${offset}`
            );

            if (!response.ok) {
                throw new Error('Failed to list images');
            }

            return await response.json();
        } catch (error) {
            console.error('List images error:', error);
            throw error;
        }
    }

    /**
     * Получить статистику
     */
    async getStatistics() {
        try {
            const response = await fetch(`${this.baseURL}/statistics`);

            if (!response.ok) {
                throw new Error('Failed to get statistics');
            }

            return await response.json();
        } catch (error) {
            console.error('Get statistics error:', error);
            throw error;
        }
    }

    /**
     * Проверить здоровье API
     */
    async healthCheck() {
        try {
            const response = await fetch(`${this.baseURL.replace('/api/v1', '')}/health`);
            return response.ok;
        } catch (error) {
            console.error('Health check error:', error);
            return false;
        }
    }

    /**
     * Опросить статус изображения с интервалом
     */
    async pollImageStatus(imageId, callback, interval = null, maxAttempts = null) {
        interval = interval || this.pollingConfig.interval;
        maxAttempts = maxAttempts || this.pollingConfig.maxAttempts;
        let attempts = 0;

        const poll = async () => {
            try {
                attempts++;
                const status = await this.getImageStatus(imageId);
                
                callback(status);

                // Если обработка завершена или произошла ошибка, остановить опрос
                if (status.status === 'completed' || status.status === 'failed') {
                    return status;
                }

                // Если превышено максимальное количество попыток
                if (attempts >= maxAttempts) {
                    throw new Error('Polling timeout');
                }

                // Продолжить опрос
                await new Promise(resolve => setTimeout(resolve, interval));
                return poll();
            } catch (error) {
                console.error('Polling error:', error);
                throw error;
            }
        };

        return poll();
    }
}

// Экспорт API
const api = new ImageProcessorAPI();

