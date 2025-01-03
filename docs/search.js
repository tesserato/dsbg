document.addEventListener('DOMContentLoaded', function () {
    const searchInput = document.getElementById('search-input');
    const searchResults = document.getElementById('search-results');
    let fuse;

    fetch('search_index.json')
        .then(response => response.json())
        .then(searchIndex => {
            const options = {
                includeScore: true,
                findAllMatches: true,
                includeMatches: true,
                ignoreLocation: true,
                minMatchCharLength: 2,
                useExtendedSearch: true,
                threshold: 0.35,
                distance: 10,
                keys: ['title', 'content', 'description', 'tags']
            };
            fuse = new Fuse(searchIndex, options);
        });


    searchInput.addEventListener('keyup', function (event) {
        if (fuse) {
            const searchTerm = event.target.value;
            const results = fuse.search(searchTerm);

            let resultsHTML = '';
            if (results.length === 0) {
                resultsHTML = '<li>No results found.</li>';
            } else {
                results.forEach(result => {
                    const article = result.item;
                    resultsHTML += `<li>${(1.0 - result.score).toFixed(2)} <a href="${article.url}">${article.title}</a></li>`;
                });
            }
            searchResults.innerHTML = resultsHTML;
        }

    });

    searchResults.addEventListener('mousedown', function (event) {
        event.preventDefault();
    });

    searchInput.addEventListener('blur', function (event) {
        searchInput.value = '';
        searchResults.innerHTML = '';
    });
});