const form = document.getElementById('shortenForm');
const submitBtn = document.getElementById('submitBtn');
const result = document.getElementById('result');
const error = document.getElementById('error');
const shortUrlInput = document.getElementById('shortUrlInput');

form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const originalUrl = document.getElementById('originalUrl').value;
    const customCode = document.getElementById('customCode').value;

    // Скрываем предыдущие результаты
    result.classList.remove('show');
    error.classList.remove('show');

    // Показываем загрузку
    const originalBtnText = submitBtn.innerHTML;
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span class="loading"></span>';

    try {
        const payload = { original_url: originalUrl };
        if (customCode) {
            payload.custom_code = customCode;
        }

        const response = await fetch('/api/v1/urls', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload)
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Ошибка создания ссылки');
        }

        // Показываем результат
        shortUrlInput.value = data.short_url;
        document.getElementById('urlId').textContent = data.id;
        document.getElementById('urlCode').textContent = data.short_code;
        document.getElementById('urlCreated').textContent = new Date(data.created_at).toLocaleString('ru-RU');
        document.getElementById('urlClicks').textContent = data.clicks_count;
        result.classList.add('show');

        // Очищаем форму
        form.reset();

    } catch (err) {
        error.textContent = err.message;
        error.classList.add('show');
    } finally {
        submitBtn.disabled = false;
        submitBtn.innerHTML = originalBtnText;
    }
});

async function copyToClipboard() {
    const btn = event.target;
    const url = shortUrlInput.value;

    try {
        await navigator.clipboard.writeText(url);
        const originalText = btn.textContent;
        btn.textContent = '✓ Скопировано';
        btn.classList.add('copied');

        setTimeout(() => {
            btn.textContent = originalText;
            btn.classList.remove('copied');
        }, 2000);
    } catch (err) {
        // Fallback для старых браузеров
        shortUrlInput.select();
        document.execCommand('copy');
        btn.textContent = '✓ Скопировано';
    }
}
