function initProgressBar() {
    console.log("Checking player state...");
    
    // Always clear the old interval first to prevent memory leaks
    if (window.musicInterval) {
        clearInterval(window.musicInterval);
        window.musicInterval = null;
    }

    const container = document.getElementById('progress-container');
    
    // If the container doesn't exist (Not Streaming mode), we just stop here.
    if (!container) {
        console.log("Not streaming - Timer cleared.");
        return;
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
            const minutes = Math.floor(currentMs / 60000);
            const seconds = Math.floor((currentMs % 60000) / 1000);
            timeDisplay.innerText = minutes + ":" + (seconds < 10 ? '0' : '') + seconds;
        }
    };

    updateUI();
    window.musicInterval = setInterval(updateUI, 1000);    
}