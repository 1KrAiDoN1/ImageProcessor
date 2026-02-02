/**
 * UI Module - управление интерфейсом
 */

const UI = {
    /**
     * Показать toast уведомление
     */
    showToast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        
        const icon = type === 'success' ? 'fa-check-circle' : 
                     type === 'error' ? 'fa-exclamation-circle' : 
                     'fa-info-circle';
        
        toast.innerHTML = `
            <i class="fas ${icon}"></i>
            <span>${message}</span>
        `;
        
        container.appendChild(toast);
        
        // Автоматически убрать через 5 секунд
        setTimeout(() => {
            toast.style.animation = 'slideIn 0.3s ease-out reverse';
            setTimeout(() => toast.remove(), 300);
        }, 5000);
    },

    /**
     * Переключить секцию
     */
    switchSection(sectionId) {
        // Скрыть все секции
        document.querySelectorAll('.section').forEach(section => {
            section.classList.remove('active');
        });
        
        // Показать выбранную секцию
        const targetSection = document.getElementById(`${sectionId}-section`);
        if (targetSection) {
            targetSection.classList.add('active');
        }
        
        // Обновить активную ссылку в навигации
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
            if (link.dataset.section === sectionId) {
                link.classList.add('active');
            }
        });
    },

    /**
     * Показать/скрыть элемент
     */
    toggleElement(element, show) {
        if (show) {
            element.classList.remove('hidden');
        } else {
            element.classList.add('hidden');
        }
    },

    /**
     * Обновить прогресс-бар
     */
    updateProgress(percentage, text, status) {
        const progressFill = document.getElementById('progressFill');
        const progressText = document.getElementById('progressText');
        const progressPercentage = document.getElementById('progressPercentage');
        const progressStatus = document.getElementById('progressStatus');
        
        if (progressFill) {
            progressFill.style.width = `${percentage}%`;
        }
        
        if (progressText && text) {
            progressText.textContent = text;
        }
        
        if (progressPercentage) {
            progressPercentage.textContent = `${percentage}%`;
        }
        
        if (progressStatus && status) {
            progressStatus.textContent = status;
            progressStatus.className = '';
            
            // Установить цвет в зависимости от статуса
            if (status === 'completed') {
                progressStatus.style.background = 'var(--success)';
            } else if (status === 'failed') {
                progressStatus.style.background = 'var(--danger)';
            } else if (status === 'processing') {
                progressStatus.style.background = 'var(--warning)';
            } else {
                progressStatus.style.background = 'var(--info)';
            }
        }
    },

    /**
     * Создать карточку результата
     */
    createResultCard(imageId, operation, operationName) {
        const card = document.createElement('div');
        card.className = 'result-card';
        card.dataset.imageId = imageId;
        card.dataset.operation = operation;
        
        const imageURL = api.getImageURL(imageId, operation);
        
        card.innerHTML = `
            <img src="${imageURL}" alt="${operationName}" class="result-image" 
                 onerror="this.src='data:image/svg+xml,%3Csvg xmlns=\\'http://www.w3.org/2000/svg\\' width=\\'200\\' height=\\'200\\'%3E%3Crect fill=\\'%23ddd\\' width=\\'200\\' height=\\'200\\'/%3E%3Ctext fill=\\'%23999\\' x=\\'50%25\\' y=\\'50%25\\' text-anchor=\\'middle\\' dy=\\'.3em\\'%3EИзображение недоступно%3C/text%3E%3C/svg%3E'">
            <div class="result-info">
                <h4>${operationName}</h4>
                <p>ID: ${imageId.substring(0, 8)}...</p>
                <span class="result-badge completed">Готово</span>
            </div>
        `;
        
        // Обработчик клика для открытия модального окна
        card.addEventListener('click', () => {
            UI.openImageModal(imageId, operation);
        });
        
        return card;
    },

    /**
     * Отобразить результаты обработки
     */
    displayResults(imageId, operations) {
        const resultsGrid = document.getElementById('resultsGrid');
        resultsGrid.innerHTML = '';
        
        // Оригинал
        const originalCard = UI.createResultCard(imageId, 'original', 'Оригинал');
        resultsGrid.appendChild(originalCard);
        
        // Обработанные версии
        operations.forEach(op => {
            const operationNames = {
                'thumbnail': 'Миниатюра',
                'resize': 'Измененный размер',
                'watermark': 'С водяным знаком'
            };
            
            const card = UI.createResultCard(
                imageId, 
                op.type, 
                operationNames[op.type] || op.type
            );
            resultsGrid.appendChild(card);
        });
    },

    /**
     * Открыть модальное окно с изображением
     */
    openImageModal(imageId, operation = 'original') {
        const modal = document.getElementById('imageModal');
        const modalImage = document.getElementById('modalImage');
        const downloadLink = document.getElementById('downloadLink');
        const deleteBtn = document.getElementById('deleteImageBtn');
        
        const imageURL = api.getImageURL(imageId, operation);
        
        modalImage.src = imageURL;
        downloadLink.href = imageURL;
        downloadLink.download = `${imageId}_${operation}.jpg`;
        
        // Установить обработчик удаления
        deleteBtn.onclick = async () => {
            if (confirm('Вы уверены, что хотите удалить это изображение?')) {
                try {
                    await api.deleteImage(imageId);
                    UI.showToast('Изображение успешно удалено', 'success');
                    UI.closeImageModal();
                    // Обновить галерею
                    if (document.getElementById('gallery-section').classList.contains('active')) {
                        await loadGallery();
                    }
                } catch (error) {
                    UI.showToast('Ошибка при удалении изображения', 'error');
                }
            }
        };
        
        modal.classList.remove('hidden');
    },

    /**
     * Закрыть модальное окно
     */
    closeImageModal() {
        const modal = document.getElementById('imageModal');
        modal.classList.add('hidden');
    },

    /**
     * Создать элемент галереи
     */
    createGalleryItem(image) {
        const item = document.createElement('div');
        item.className = 'gallery-item';
        item.dataset.imageId = image.id;
        item.dataset.status = image.status;
        
        const imageURL = api.getImageURL(image.id, 'thumbnail');
        
        item.innerHTML = `
            <img src="${imageURL}" alt="${image.filename}"
                 onerror="this.src='data:image/svg+xml,%3Csvg xmlns=\\'http://www.w3.org/2000/svg\\' width=\\'250\\' height=\\'250\\'%3E%3Crect fill=\\'%23ddd\\' width=\\'250\\' height=\\'250\\'/%3E%3Ctext fill=\\'%23999\\' x=\\'50%25\\' y=\\'50%25\\' text-anchor=\\'middle\\' dy=\\'.3em\\'%3E${image.status}%3C/text%3E%3C/svg%3E'">
            <div class="gallery-item-overlay">
                <h4>${image.filename}</h4>
                <p>${UI.formatFileSize(image.size)}</p>
                <p>${UI.formatDate(image.created_at)}</p>
            </div>
        `;
        
        item.addEventListener('click', () => {
            UI.openImageModal(image.id, 'original');
        });
        
        return item;
    },

    /**
     * Форматировать размер файла
     */
    formatFileSize(bytes) {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    },

    /**
     * Форматировать дату
     */
    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleString('ru-RU', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    },

    /**
     * Отобразить загрузчик
     */
    showLoader(container) {
        container.innerHTML = `
            <div class="loading-spinner">
                <i class="fas fa-spinner fa-spin"></i>
                <p>Загрузка...</p>
            </div>
        `;
    },

    /**
     * Показать пустое состояние
     */
    showEmptyState(container, message, icon = 'fa-image') {
        container.innerHTML = `
            <div class="loading-spinner">
                <i class="fas ${icon}" style="animation: none;"></i>
                <p>${message}</p>
            </div>
        `;
    },

    /**
     * Создать таблицу статистики операций
     */
    createOperationsTable(operations) {
        if (!operations || operations.length === 0) {
            return '<p style="text-align: center; color: var(--gray-dark);">Нет данных</p>';
        }
        
        let html = `
            <table>
                <thead>
                    <tr>
                        <th>Операция</th>
                        <th>Всего</th>
                        <th>Успешно</th>
                        <th>Ошибок</th>
                        <th>Среднее время (мс)</th>
                        <th>Успешность</th>
                    </tr>
                </thead>
                <tbody>
        `;
        
        operations.forEach(op => {
            const successRate = op.total_count > 0 
                ? ((op.success_count / op.total_count) * 100).toFixed(1) 
                : 0;
            
            html += `
                <tr>
                    <td><strong>${UI.getOperationName(op.operation_type)}</strong></td>
                    <td>${op.total_count}</td>
                    <td style="color: var(--success);">${op.success_count}</td>
                    <td style="color: var(--danger);">${op.failure_count}</td>
                    <td>${op.average_processing_time_ms.toFixed(1)}</td>
                    <td>
                        <div style="display: flex; align-items: center; gap: 0.5rem;">
                            <div style="flex: 1; height: 8px; background: var(--gray); border-radius: 4px; overflow: hidden;">
                                <div style="width: ${successRate}%; height: 100%; background: var(--success);"></div>
                            </div>
                            <span>${successRate}%</span>
                        </div>
                    </td>
                </tr>
            `;
        });
        
        html += '</tbody></table>';
        return html;
    },

    /**
     * Получить русское название операции
     */
    getOperationName(type) {
        const names = {
            'thumbnail': 'Миниатюра',
            'resize': 'Изменение размера',
            'watermark': 'Водяной знак',
            'crop': 'Обрезка',
            'rotate': 'Поворот',
            'flip': 'Отражение',
            'grayscale': 'Черно-белое'
        };
        return names[type] || type;
    }
};

