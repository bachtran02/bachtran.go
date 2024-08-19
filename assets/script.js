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
    console.log("scoreboard successfully updated")
}

function scoreboardError() {
    document.querySelector("#scoreboard").innerHTML = `<span class="error">Error fetching scoreboard data</span>`;
}

async function loadMusic() {
    let response;
    try {
        response = await fetch(`/api/music`, {
            method: "GET"
        });
    } catch (e) {
        console.error("error fetching music:", e);
        musicError();
        return;
    }

    if (!response.ok) {
        console.error("error fetching music:", response);
        musicError();
        return;
    }

    document.querySelector("#music").innerHTML = await response.text();
}

function musicError() {
    document.querySelector("#music").innerHTML = `<span class="error">Error fetching music data</span>`;
}


document.addEventListener('DOMContentLoaded', async () => {
    await loadScoreboard();
    setInterval(loadScoreboard, 1000 * 10);
    
    await loadMusic();
    setInterval(loadMusic, 1000 * 5);
});