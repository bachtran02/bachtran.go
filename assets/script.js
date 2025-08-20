async function loadScoreboard() {
    let response;
    try {
        response = await fetch(`/api/scoreboard`, {
            method: "GET"
        });
    } catch (e) {
        console.error("error fetching scoreboard:", e);
        scoreboardError();
        return;
    }

    if (!response.ok) {
        console.error("error fetching scoreboard:", response);
        scoreboardError();
        return;
    }
    document.querySelector("#scoreboard").innerHTML = await response.text();
    convertMatchTimesToUserTimezone(); // Call the new function here
}

function convertMatchTimesToUserTimezone() {
    const matchTimeElements = document.querySelectorAll('[data-original-time]');

    matchTimeElements.forEach(element => {
        const originalTimeStr = element.getAttribute('data-original-time');
        if (originalTimeStr) {
            try {
                const date = new Date(originalTimeStr);
                const userTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

                const formattedDate = new Intl.DateTimeFormat(
                    undefined, { month: 'short', day: 'numeric', timeZone: userTimeZone }).format(date);
                const formattedTime = new Intl.DateTimeFormat(
                    undefined, { hour: '2-digit', minute: '2-digit', hour12: false, timeZone: userTimeZone }).format(date);

                element.innerHTML = `${formattedDate} at <strong>${formattedTime}</strong>`;

            } catch (e) {
                console.error('Error converting time:', e);
            }
        }
    });
}

function scoreboardError() {
    document.querySelector("#scoreboard").innerHTML = `<span class="error">Error fetching scoreboard data</span>`;
}

let currentTrack = null;
let animationFrameId = null;
let lastUpdateTime = 0;

// Helper function to format milliseconds to MM:SS or HH:MM:SS
function formatTime(ms) {
    const totalSeconds = Math.floor(ms / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;

    if (hours > 0) {
        return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
    } else {
        return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
    }
}

function renderProgressBar() {
    const currentTimeSpan = document.getElementById("current-time");
    const totalDurationSpan = document.getElementById("total-duration");

    if (!currentTrack || !currentTrack.is_playing) {
        if (animationFrameId) {
            cancelAnimationFrame(animationFrameId);
            animationFrameId = null;
        }
        // Reset time displays when not playing
        if (currentTimeSpan) currentTimeSpan.innerText = "00:00";
        if (totalDurationSpan) totalDurationSpan.innerText = "00:00";
        return;
    }

    if (currentTrack.is_paused) {
        if (animationFrameId) {
            cancelAnimationFrame(animationFrameId);
            animationFrameId = null;
        }
        return;
    }

    if (currentTrack.is_stream) {
        const progressBar = document.getElementById("progress-bar");
        if (progressBar) {
            progressBar.style.width = `100%`;
        }
        if (currentTimeSpan) {
            currentTimeSpan.innerText = "";
        }
        if (totalDurationSpan) {
            totalDurationSpan.innerText = "LIVE";
        }
        return;
    }

    const now = Date.now();
    const elapsed = now - lastUpdateTime;
    let estimatedPosition = currentTrack.position_ms + elapsed;

    // Ensure estimated position doesn't exceed duration
    if (estimatedPosition > currentTrack.duration_ms) {
        estimatedPosition = currentTrack.duration_ms;
    }

    const progressBar = document.getElementById("progress-bar");
    if (progressBar) {
        const progress = (estimatedPosition / currentTrack.duration_ms) * 100;
        progressBar.style.width = `${progress}%`;
    }

    // Update time displays
    if (currentTimeSpan) {
        currentTimeSpan.innerText = formatTime(estimatedPosition);
    }
    if (totalDurationSpan) {
        totalDurationSpan.innerText = formatTime(currentTrack.duration_ms);
    }
    animationFrameId = requestAnimationFrame(renderProgressBar);
}

// --- Autoscroll Logic ---
const speed = 50;               // px/sec while hovered
const returnMs = 1500;          // duration to scroll back on mouseleave (ms)

let rafScroll = null;
let rafReturn = null;
let last, dir = 1;

function scrollLoop(ts) {
    const el = document.getElementById('track-title');
    if (last == null) last = ts;
    const dt = (ts - last) / 1000; // seconds
    last = ts;

    const end = Math.max(0, el.scrollWidth - el.clientWidth);
    el.scrollLeft = Math.min(end, Math.max(0, el.scrollLeft + dir * speed * dt));

    if (el.scrollLeft >= end) {
        cancelAnimationFrame(rafScroll);
        rafScroll = null;
        return;
    }
    rafScroll = requestAnimationFrame(scrollLoop);
}

function startHover() {
    const el = document.getElementById('track-title');
    if (el.scrollWidth <= el.clientWidth) return; // nothing to scroll
    // cancel return animation if it was running
    if (rafReturn) { cancelAnimationFrame(rafReturn); rafReturn = null; }
    if (!rafScroll) { last = undefined; rafScroll = requestAnimationFrame(scrollLoop); }
}

function easeOutCubic(t){ return 1 - Math.pow(1 - t, 3); }

function returnToStart() {
    const el = document.getElementById('track-title');
    const from = el.scrollLeft;
    if (from <= 0) return;

    const t0 = performance.now();

    function step(t) {
        const p = Math.min((t - t0) / returnMs, 1);
        const eased = easeOutCubic(p);
        el.scrollLeft = from * (1 - eased); // animate back to 0
        if (p < 1) rafReturn = requestAnimationFrame(step);
        else rafReturn = null;
    }
    rafReturn = requestAnimationFrame(step);
}

function endHover() {
    if (rafScroll) { cancelAnimationFrame(rafScroll); rafScroll = null; }
    returnToStart();
}

function connectWebSocket() {
    // TODO: update this
    const ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = () => {
        console.log("WebSocket: Connected to music websocket");
    };

    ws.onmessage = (event) => {
        // console.log("WebSocket: Received message:", event.data);
        const data = JSON.parse(event.data);
        const musicPlayerContent = document.getElementById("music-player-content");
        const notStreamingMessage = document.getElementById("not-streaming-message");
        const trackTitleLink = document.getElementById("track-title"); // Get reference here

        if (data.is_playing) {
            musicPlayerContent.style.display = "block";
            notStreamingMessage.style.display = "none";

            // Update currentTrack state based on incoming data.track
            currentTrack = {
                is_playing: data.is_playing,
                is_paused: data.is_paused,
                title: data.track.title,
                artist: data.track.author,
                album_art_url: data.track.artworkUrl,
                position_ms: data.track.position,
                duration_ms: data.track.length,
                is_stream: data.track.isStream,
            };
            lastUpdateTime = Date.now(); // Reset last update time

            document.getElementById("track-artwork").src = currentTrack.album_art_url;
            
            if (data.track.uri) { // Use data.track.uri for the link
                trackTitleLink.href = data.track.uri;
            } else {
                trackTitleLink.removeAttribute("href"); // Remove link if no URI
            }
            trackTitleLink.innerText = currentTrack.title;
            document.getElementById("track-artist").innerText = currentTrack.artist;

            // Immediately update progress bar to the received position
            const progressBar = document.getElementById("progress-bar");
            if (progressBar && currentTrack.is_stream) {
                progressBar.style.width = `100%`; // For streams, show full width
            } else if (progressBar && currentTrack.duration_ms > 0) {
                const initialProgress = (currentTrack.position_ms / currentTrack.duration_ms) * 100;
                progressBar.style.width = `${initialProgress}%`;
            } else if (progressBar) {
                progressBar.style.width = `0%`; // Reset if duration is zero or not playing
            }

            // Initial update of time displays
            const currentTimeSpan = document.getElementById("current-time");
            const totalDurationSpan = document.getElementById("total-duration");
            if (currentTrack.is_stream) {
                if (currentTimeSpan) currentTimeSpan.innerText = "";
                if (totalDurationSpan) totalDurationSpan.innerText = "LIVE";
            } else {
                if (currentTimeSpan) currentTimeSpan.innerText = formatTime(currentTrack.position_ms);
                if (totalDurationSpan) totalDurationSpan.innerText = formatTime(currentTrack.duration_ms);
            }

            // Autoscroll Event Listeners
            trackTitleLink.addEventListener('mouseenter', startHover);
            trackTitleLink.addEventListener('mouseleave', endHover);

            // Start animation if not already running
            if (!animationFrameId) {
                animationFrameId = requestAnimationFrame(renderProgressBar);
            }

        } else {
            musicPlayerContent.style.display = "none";
            notStreamingMessage.style.display = "block";
            currentTrack = null; // Clear current track data
            if (animationFrameId) {
                cancelAnimationFrame(animationFrameId);
                animationFrameId = null;
            }
            // Reset progress bar and time displays when not streaming
            const progressBar = document.getElementById("progress-bar");
            if (progressBar) {
                progressBar.style.width = `0%`;
            }
            const currentTimeSpan = document.getElementById("current-time");
            const totalDurationSpan = document.getElementById("total-duration");
            if (currentTimeSpan) currentTimeSpan.innerText = "00:00";
            if (totalDurationSpan) totalDurationSpan.innerText = "00:00";
        }
    };

    ws.onclose = (event) => {
        // console.log("WebSocket: Music websocket closed, code:", event.code, "reason:", event.reason);
        currentTrack = null;
        if (animationFrameId) {
            cancelAnimationFrame(animationFrameId);
            animationFrameId = null;
        }
        const musicPlayerContent = document.getElementById("music-player-content");
        const notStreamingMessage = document.getElementById("not-streaming-message");
        if (musicPlayerContent) musicPlayerContent.style.display = "none";
        if (notStreamingMessage) notStreamingMessage.style.display = "block";
        const progressBar = document.getElementById("progress-bar");
        if (progressBar) progressBar.style.width = `0%`;
        const currentTimeSpan = document.getElementById("current-time");
        const totalDurationSpan = document.getElementById("total-duration");
        if (currentTimeSpan) currentTimeSpan.innerText = "00:00";
        if (totalDurationSpan) totalDurationSpan.innerText = "00:00";

        // Attempt to reconnect after a delay
        setTimeout(connectWebSocket, 3000); // Increased delay for better stability
    };

    ws.onerror = (err) => {
        console.error("WebSocket: Music websocket error:", err);
        ws.close(); // Close to trigger onclose and reconnect logic
    };
}

document.addEventListener('DOMContentLoaded', async () => {
    await loadScoreboard();
    setInterval(loadScoreboard, 1000 * 10);

    connectWebSocket();
});