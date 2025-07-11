let currentPage = 1;
const postsPerPage = 10;

export async function loadPosts() {
    try {
        const response = await fetch(`/api/posts?page=${currentPage}&limit=${postsPerPage}`);
        const posts = await response.json();

        if (posts.length === 0) {
            document.getElementById('load-more').disabled = true;
            return;
        }

        renderPosts(posts);
        currentPage++;
    } catch (error) {
        showError('Failed to load posts');
    }
}

export function loadMorePosts() {
    loadPosts();
}

function renderPosts(posts) {
    const container = document.getElementById('posts-list');

    posts.forEach(post => {
        const postElement = document.createElement('div');
        postElement.classList.add('post');
        postElement.innerHTML = `
            <div class="post-header">
                <span class="author">${post.author}</span>
                <span class="time">${new Date(post.createdAt).toLocaleString()}</span>
            </div>
            <h3 class="post-title">${post.title}</h3>
            <div class="post-content">${post.content}</div>
            <div class="post-categories">${post.categories.join(', ')}</div>
            <div class="post-actions">
                <button class="like-btn" data-post-id="${post.id}">Like</button>
                <button class="comment-btn" data-post-id="${post.id}">Comment</button>
            </div>
            <div class="comments-section hidden" id="comments-${post.id}"></div>
        `;

        container.appendChild(postElement);

        // Add event listeners
        postElement.querySelector('.like-btn').addEventListener('click', handleLike);
        postElement.querySelector('.comment-btn').addEventListener('click', toggleComments);
    });
}

function handleLike(event) {
    const postId = event.target.dataset.postId;
    fetch(`/api/posts/${postId}/like`, { method: 'POST' })
        .then(response => {
            if (response.ok) {
                event.target.textContent = 'Liked!';
            }
        });
}

function toggleComments(event) {
    const postId = event.target.dataset.postId;
    const commentsSection = document.getElementById(`comments-${postId}`);

    if (commentsSection.classList.contains('hidden')) {
        loadComments(postId);
        commentsSection.classList.remove('hidden');
    } else {
        commentsSection.classList.add('hidden');
    }
}

async function loadComments(postId) {
    try {
        const response = await fetch(`/api/posts/${postId}/comments`);
        const comments = await response.json();
        renderComments(postId, comments);
    } catch (error) {
        showError('Failed to load comments');
    }
}

function renderComments(postId, comments) {
    const container = document.getElementById(`comments-${postId}`);
    container.innerHTML = '';

    comments.forEach(comment => {
        const commentElement = document.createElement('div');
        commentElement.classList.add('comment');
        commentElement.innerHTML = `
            <div class="comment-header">
                <span class="author">${comment.author}</span>
                <span class="time">${new Date(comment.createdAt).toLocaleString()}</span>
            </div>
            <div class="comment-content">${comment.content}</div>
        `;
        container.appendChild(commentElement);
    });

    // Add comment form
    const commentForm = document.createElement('div');
    commentForm.classList.add('comment-form');
    commentForm.innerHTML = `
        <textarea placeholder="Write a comment..." id="comment-input-${postId}"></textarea>
        <button class="submit-comment" data-post-id="${postId}">Submit</button>
    `;
    container.appendChild(commentForm);

    commentForm.querySelector('.submit-comment').addEventListener('click', submitComment);
}

function submitComment(event) {
    const postId = event.target.dataset.postId;
    const content = document.getElementById(`comment-input-${postId}`).value.trim();

    if (content) {
        fetch(`/api/posts/${postId}/comments`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ content })
        })
            .then(response => {
                if (response.ok) {
                    loadComments(postId);
                }
            });
    }
}