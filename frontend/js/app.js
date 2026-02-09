/**
 * Main Application Logic
 */

// Глобальные переменные
let selectedFile = null;
let currentImageId = null;
let currentPage = 0;
const PAGE_SIZE = 12;

/**
 * Инициализация приложения
 */
document.addEventListener('DOMContentLoaded', async () => {
    initializeNavigation();
    initializeUploadSection();
    initializeOperations();
    initializeModal();
    
    // Проверка здоровья API
    const isHealthy = await api.healthCheck();
    if (!isHealthy) {
        UI.showToast('Не удается подключиться к серверу', 'error');
    }
    
    // Загрузить начальные данные
    await loadStatistics();
});

/**
 * Инициализация навигации
 */
function initializeNavigation() {
    document.querySelectorAll('.nav-link').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const section = link.dataset.section;
            UI.switchSection(section);
            
            // Загрузить данные для секции
            if (section === 'gallery') {
                loadGallery();
            } else if (section === 'stats') {
                loadStatistics();
            }
        });
    });
}

/**
 * Инициализация секции загрузки
 */
function initializeUploadSection() {
    const uploadArea = document.getElementById('uploadArea');
    const fileInput = document.getElementById('fileInput');
    const selectFileBtn = document.getElementById('selectFileBtn');
    const removeImageBtn = document.getElementById('removeImageBtn');
    const uploadBtn = document.getElementById('uploadBtn');
    const newUploadBtn = document.getElementById('newUploadBtn');
    
    // Обработчик клика на кнопку выбора файла
    selectFileBtn.addEventListener('click', () => {
        fileInput.click();
    });
    
    // Обработчик выбора файла
    fileInput.addEventListener('change', (e) => {
        const file = e.target.files[0];
        if (file) {
            handleFileSelect(file);
        }
    });
    
    // Drag & Drop
    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('drag-over');
    });
    
    uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('drag-over');
    });
    
    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('drag-over');
        
        const file = e.dataTransfer.files[0];
        if (file && file.type.startsWith('image/')) {
            handleFileSelect(file);
        } else {
            UI.showToast('Пожалуйста, выберите изображение', 'error');
        }
    });
    
    // Обработчик удаления изображения
    removeImageBtn.addEventListener('click', () => {
        resetUploadSection();
    });
    
    // Обработчик загрузки
    uploadBtn.addEventListener('click', () => {
        handleUpload();
    });
    
    // Обработчик новой загрузки
    newUploadBtn.addEventListener('click', () => {
        resetUploadSection();
        UI.switchSection('upload');
    });
}

/**
 * Обработка выбора файла
 */
function handleFileSelect(file) {
    // Проверка типа файла
    const validTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
    if (!validTypes.includes(file.type)) {
        UI.showToast('Неподдерживаемый формат файла', 'error');
        return;
    }
    
    // Проверка размера файла (32MB)
    const maxSize = 32 * 1024 * 1024;
    if (file.size > maxSize) {
        UI.showToast('Размер файла превышает 32MB', 'error');
        return;
    }
    
    selectedFile = file;
    
    // Показать превью
    const reader = new FileReader();
    reader.onload = (e) => {
        const previewImage = document.getElementById('previewImage');
        previewImage.src = e.target.result;
        
        const imageName = document.getElementById('imageName');
        imageName.textContent = file.name;
        
        const imageSize = document.getElementById('imageSize');
        imageSize.textContent = UI.formatFileSize(file.size);
        
        // Переключить области
        UI.toggleElement(document.getElementById('uploadArea'), false);
        UI.toggleElement(document.getElementById('previewArea'), true);
        UI.toggleElement(document.getElementById('operationsPanel'), true);
    };
    reader.readAsDataURL(file);
}

/**
 * Сброс секции загрузки
 */
function resetUploadSection() {
    selectedFile = null;
    currentImageId = null;
    
    document.getElementById('fileInput').value = '';
    document.getElementById('previewImage').src = '';
    
    UI.toggleElement(document.getElementById('uploadArea'), true);
    UI.toggleElement(document.getElementById('previewArea'), false);
    UI.toggleElement(document.getElementById('operationsPanel'), false);
    UI.toggleElement(document.getElementById('progressSection'), false);
    UI.toggleElement(document.getElementById('resultsSection'), false);
    
    // Сбросить чекбоксы операций
    document.querySelectorAll('.operation-checkbox').forEach(cb => {
        cb.checked = false;
    });
}

/**
 * Инициализация операций
 */
function initializeOperations() {
    // Thumbnail
    const thumbnailCheckbox = document.getElementById('opThumbnail');
    thumbnailCheckbox.addEventListener('change', () => {
        UI.toggleElement(
            document.getElementById('thumbnailParams'),
            thumbnailCheckbox.checked
        );
    });
    
    // Resize
    const resizeCheckbox = document.getElementById('opResize');
    resizeCheckbox.addEventListener('change', () => {
        UI.toggleElement(
            document.getElementById('resizeParams'),
            resizeCheckbox.checked
        );
    });
    
    // Watermark
    const watermarkCheckbox = document.getElementById('opWatermark');
    watermarkCheckbox.addEventListener('change', () => {
        UI.toggleElement(
            document.getElementById('watermarkParams'),
            watermarkCheckbox.checked
        );
    });
    
    // Обновление значения прозрачности
    const opacityRange = document.getElementById('watermarkOpacity');
    const opacityValue = document.getElementById('opacityValue');
    opacityRange.addEventListener('input', () => {
        opacityValue.textContent = opacityRange.value;
    });
}

/**
 * Получить выбранные операции
 */
function getSelectedOperations() {
    const operations = [];
    
    // Thumbnail
    if (document.getElementById('opThumbnail').checked) {
        operations.push({
            type: 'thumbnail',
            parameters: {
                size: parseInt(document.getElementById('thumbnailSize').value),
                crop_to_fit: document.getElementById('thumbnailCrop').checked
            }
        });
    }
    
    // Resize
    if (document.getElementById('opResize').checked) {
        operations.push({
            type: 'resize',
            parameters: {
                width: parseInt(document.getElementById('resizeWidth').value),
                height: parseInt(document.getElementById('resizeHeight').value),
                keep_aspect: document.getElementById('resizeKeepAspect').checked
            }
        });
    }
    
    // Watermark
    if (document.getElementById('opWatermark').checked) {
        operations.push({
            type: 'watermark',
            parameters: {
                text: document.getElementById('watermarkText').value,
                opacity: parseFloat(document.getElementById('watermarkOpacity').value),
                position: document.getElementById('watermarkPosition').value
            }
        });
    }
    
    return operations;
}

/**
 * Обработка загрузки
 */
async function handleUpload() {
    if (!selectedFile) {
        UI.showToast('Пожалуйста, выберите изображение', 'error');
        return;
    }
    
    const operations = getSelectedOperations();
    
    if (operations.length === 0) {
        UI.showToast('Пожалуйста, выберите хотя бы одну операцию', 'error');
        return;
    }
    
    // Показать прогресс
    UI.toggleElement(document.getElementById('operationsPanel'), false);
    UI.toggleElement(document.getElementById('progressSection'), true);
    UI.updateProgress(0, 'Загрузка изображения...', 'pending');
    
    try {
        // Загрузить изображение
        UI.updateProgress(30, 'Загрузка на сервер...', 'uploading');
        const uploadResult = await api.uploadImage(selectedFile, operations);
        
        currentImageId = uploadResult.id;
        
        UI.updateProgress(50, 'Обработка изображения...', 'processing');
        UI.showToast('Изображение загружено, начинается обработка', 'success');
        
        // Опрашивать статус
        await api.pollImageStatus(uploadResult.id, (status) => {
            const progress = status.progress || 50;
            UI.updateProgress(
                Math.min(progress, 95),
                `Обработка: ${status.processed_operations || 0}/${status.total_operations || operations.length}`,
                status.status
            );
        });
        
        // Завершено
        UI.updateProgress(100, 'Обработка завершена!', 'completed');
        UI.showToast('Изображение успешно обработано!', 'success');
        
        // Показать результаты
        setTimeout(() => {
            UI.toggleElement(document.getElementById('progressSection'), false);
            UI.toggleElement(document.getElementById('resultsSection'), true);
            UI.displayResults(currentImageId, operations);
        }, 200);
        
    } catch (error) {
        console.error('Upload error:', error);
        UI.updateProgress(0, 'Ошибка обработки', 'failed');
        UI.showToast('Ошибка при загрузке: ' + error.message, 'error');
        
        // Вернуться к форме
        setTimeout(() => {
            UI.toggleElement(document.getElementById('progressSection'), false);
            UI.toggleElement(document.getElementById('operationsPanel'), true);
        }, 2000);
    }
}

/**
 * Загрузка галереи
 */
async function loadGallery(filter = 'all', page = 0) {
    const galleryGrid = document.getElementById('galleryGrid');
    UI.showLoader(galleryGrid);
    
    try {
        const result = await api.listImages(PAGE_SIZE, page * PAGE_SIZE);
        
        if (!result.images || result.images.length === 0) {
            UI.showEmptyState(galleryGrid, 'Нет изображений', 'fa-images');
            return;
        }
        
        // Фильтрация
        let images = result.images;
        if (filter !== 'all') {
            images = images.filter(img => img.status === filter);
        }
        
        // Отобразить изображения
        galleryGrid.innerHTML = '';
        images.forEach(image => {
            const item = UI.createGalleryItem(image);
            galleryGrid.appendChild(item);
        });
        
        // Обновить пагинацию
        updatePagination(result.count, page);
        
    } catch (error) {
        console.error('Load gallery error:', error);
        UI.showEmptyState(galleryGrid, 'Ошибка загрузки галереи', 'fa-exclamation-triangle');
        UI.showToast('Ошибка загрузки галереи', 'error');
    }
}

/**
 * Обновление пагинации
 */
function updatePagination(total, currentPage) {
    const pagination = document.getElementById('pagination');
    const totalPages = Math.ceil(total / PAGE_SIZE);
    
    if (totalPages <= 1) {
        pagination.innerHTML = '';
        return;
    }
    
    let html = '';
    for (let i = 0; i < totalPages; i++) {
        const activeClass = i === currentPage ? 'active' : '';
        html += `
            <button class="filter-btn ${activeClass}" data-page="${i}">
                ${i + 1}
            </button>
        `;
    }
    
    pagination.innerHTML = html;
    
    // Добавить обработчики
    pagination.querySelectorAll('button').forEach(btn => {
        btn.addEventListener('click', () => {
            const page = parseInt(btn.dataset.page);
            loadGallery('all', page);
        });
    });
}

/**
 * Загрузка статистики
 */
async function loadStatistics() {
    try {
        const stats = await api.getStatistics();
        
        console.log('Statistics loaded:', stats); // Debug log
        
        // Обновить общую статистику
        const uploadedEl = document.getElementById('statUploaded');
        const processedEl = document.getElementById('statProcessed');
        const sizeEl = document.getElementById('statSize');
        const avgTimeEl = document.getElementById('statAvgTime');
        
        if (uploadedEl) {
            uploadedEl.textContent = stats.total_images_uploaded || 0;
        }
        if (processedEl) {
            processedEl.textContent = stats.total_images_processed || 0;
        }
        if (sizeEl) {
            // Используем total_data_processed_mb, если есть, иначе вычисляем из bytes
            const sizeMB = stats.total_data_processed_mb !== undefined 
                ? stats.total_data_processed_mb 
                : (stats.total_data_processed_bytes || 0) / (1024 * 1024);
            sizeEl.textContent = sizeMB.toFixed(2) + ' MB';
        }
        if (avgTimeEl) {
            avgTimeEl.textContent = (stats.average_processing_time_ms || 0).toFixed(1) + ' ms';
        }
        
        // Обновить статистику по операциям
        const operationsTable = document.getElementById('operationsTable');
        if (operationsTable && stats.operation_statistics) {
            operationsTable.innerHTML = UI.createOperationsTable(stats.operation_statistics);
        }
        
    } catch (error) {
        console.error('Load statistics error:', error);
        UI.showToast('Ошибка загрузки статистики', 'error');
    }
}

/**
 * Инициализация модального окна
 */
function initializeModal() {
    const modal = document.getElementById('imageModal');
    const modalClose = document.getElementById('modalClose');
    const modalOverlay = document.getElementById('modalOverlay');
    
    modalClose.addEventListener('click', () => {
        UI.closeImageModal();
    });
    
    modalOverlay.addEventListener('click', () => {
        UI.closeImageModal();
    });
    
    // Закрытие по ESC
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && !modal.classList.contains('hidden')) {
            UI.closeImageModal();
        }
    });
}

/**
 * Инициализация фильтров галереи
 */
document.addEventListener('DOMContentLoaded', () => {
    // Фильтры
    document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.filter-btn').forEach(b => {
                b.classList.remove('active');
            });
            btn.classList.add('active');
            
            const filter = btn.dataset.filter;
            if (filter) {
                loadGallery(filter, 0);
            }
        });
    });
    
    // Поиск
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        let searchTimeout;
        searchInput.addEventListener('input', () => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                const query = searchInput.value.toLowerCase();
                filterGalleryBySearch(query);
            }, 300);
        });
    }
    
    // Кнопка обновления статистики
    const refreshStatsBtn = document.getElementById('refreshStatsBtn');
    if (refreshStatsBtn) {
        refreshStatsBtn.addEventListener('click', () => {
            loadStatistics();
            UI.showToast('Статистика обновлена', 'success');
        });
    }
});

/**
 * Фильтрация галереи по поиску
 */
function filterGalleryBySearch(query) {
    const items = document.querySelectorAll('.gallery-item');
    
    items.forEach(item => {
        const overlay = item.querySelector('.gallery-item-overlay h4');
        if (overlay) {
            const filename = overlay.textContent.toLowerCase();
            if (filename.includes(query)) {
                item.style.display = '';
            } else {
                item.style.display = 'none';
            }
        }
    });
}

// Автообновление статистики каждые 30 секунд
setInterval(() => {
    if (document.getElementById('stats-section').classList.contains('active')) {
        loadStatistics();
    }
}, 30000);