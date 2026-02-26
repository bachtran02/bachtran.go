function initProgressBar() {    
    // Clean up old interval to prevent memory leaks
    if (window.musicInterval) {
        clearInterval(window.musicInterval);
        window.musicInterval = null;
    }

    const container = document.getElementById('progress-container');
    if (!container) {
        return; // Container doesn't exist
    }

    const isStream = container.getAttribute('data-is-stream') === 'true';
    if (isStream) {
        return; // No need to start interval for streams
    }

    let currentMs = parseInt(container.getAttribute('data-current'));
    const totalMs = parseInt(container.getAttribute('data-total'));
    const bar = document.getElementById('progress-bar');
    const timeDisplay = document.getElementById('current-time-display');

    const updateUI = () => {
        if (currentMs >= totalMs) {
            clearInterval(window.musicInterval);
            return;
        }

        currentMs += 1000;
        const percent = (currentMs / totalMs) * 100;
        if (bar) bar.style.width = percent + '%';

        if (timeDisplay) {
            const hours = Math.floor(currentMs / 3600000);
            const minutes = Math.floor((currentMs % 3600000) / 60000);
            const seconds = Math.floor((currentMs % 60000) / 1000);

            let timeString = "";

            if (hours > 0) {
                timeString = hours + ":" + (minutes < 10 ? '0' : '') + minutes + ":" + (seconds < 10 ? '0' : '') + seconds;
            } else {
                timeString = minutes + ":" + (seconds < 10 ? '0' : '') + seconds;
            }
            timeDisplay.innerText = timeString;
        }
    };
    updateUI();
    window.musicInterval = setInterval(updateUI, 1000);    
}