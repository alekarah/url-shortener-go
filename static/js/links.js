let currentPage = 0;
const pageSize = 20;

async function loadURLs() {
    const loading = document.getElementById('loading');
    const error = document.getElementById('error');
    const tableBody = document.getElementById('urlsTable');
    const emptyState = document.getElementById('emptyState');
    const prevBtn = document.getElementById('prevBtn');
    const nextBtn = document.getElementById('nextBtn');

    loading.style.display = 'block';
    error.style.display = 'none';
    emptyState.style.display = 'none';

    try {
        const offset = currentPage * pageSize;
        const response = await fetch(`/api/v1/urls?limit=${pageSize}&offset=${offset}`);

        if (!response.ok) {
            throw new Error('Ошибка загрузки данных');
        }

        const urls = await response.json();

        loading.style.display = 'none';

        if (!urls || urls.length === 0) {
            if (currentPage === 0) {
                emptyState.style.display = 'block';
            }
            nextBtn.disabled = true;
            return;
        }

        tableBody.innerHTML = '';

        urls.forEach(url => {
            const row = document.createElement('tr');

            const createdDate = new Date(url.created_at).toLocaleString('ru-RU');

            row.innerHTML = `
                <td><a href="${url.short_url}" class="url-link" target="_blank">${url.short_code}</a></td>
                <td><div class="original-url" title="${url.original_url}">${url.original_url}</div></td>
                <td><span class="clicks-badge">${url.clicks_count}</span></td>
                <td class="date">${createdDate}</td>
            `;

            tableBody.appendChild(row);
        });

        // Обновляем статистику
        updateStats(urls);

        // Управление кнопками пагинации
        prevBtn.disabled = currentPage === 0;
        nextBtn.disabled = urls.length < pageSize;

    } catch (err) {
        loading.style.display = 'none';
        error.textContent = err.message;
        error.style.display = 'block';
    }
}

function updateStats(urls) {
    const totalUrls = document.getElementById('totalUrls');
    const totalClicks = document.getElementById('totalClicks');

    // Текущая страница
    const startIndex = currentPage * pageSize + 1;
    const endIndex = currentPage * pageSize + urls.length;
    totalUrls.textContent = `${startIndex}-${endIndex}`;

    // Сумма кликов на текущей странице
    const clicks = urls.reduce((sum, url) => sum + url.clicks_count, 0);
    totalClicks.textContent = clicks;
}

function nextPage() {
    currentPage++;
    loadURLs();
}

function prevPage() {
    if (currentPage > 0) {
        currentPage--;
        loadURLs();
    }
}

// Загружаем данные при загрузке страницы
document.addEventListener('DOMContentLoaded', loadURLs);
