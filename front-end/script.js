document.addEventListener('DOMContentLoaded', () => {
    // 获取页面元素
    const submitBtn = document.getElementById('submitBtn');
    const nameInput = document.getElementById('nameInput');
    const commentInput = document.getElementById('commentInput');
    const commentsContainer = document.getElementById('commentsContainer');
    // 一页最大评论容量，可以随时修改
    const maxCommentsPerPage = 5;
    // 后端接口地址, 这里用了本地的8080端口
    const apiUrl="http://localhost:8080/comment";
    let currentPage = 1;
    let totalComments = 0;

    // 定时更新评论区内容
    setInterval(fetchComments, 30000); // 每30秒更新一次评论区内容

    // 获取评论的函数，异步实现，严格按照接口文档，参数直接放在Url里
    async function fetchComments() {
        const response = await fetch(`${apiUrl}/get?page=${currentPage}&size=${maxCommentsPerPage}`);
        const result = await response.json();

        if (result.code === 0) {
            totalComments = result.data.total;
            //console.log(result);
            renderComments(result.data.comments);
        } else {
            alert('Failed to fetch comments: ' + result.msg);
        }
    }

    // 渲染评论到页面上
    function renderComments(comments) {
        // 先清空原有的评论，再遍历Comments数组
        commentsContainer.innerHTML = '';
        comments.forEach((comment) => {
            // 每一个Comment的html都是一致的，直接用innerHTML返回，再指定类为comment用css控制样式，最后加上删除按钮
            const commentElement = document.createElement('div');
            commentElement.className = 'comment';

            commentElement.innerHTML = `
                <div class="comment-nameandcontent">
                    <h3>${comment.name}</h3>
                    <div class="content">${comment.content}</div>
                </div>
                <div class="delete-btn">
                    <button class="delete" data-id="${comment.id}">删除</button>
                </div>
            `;

            commentElement.querySelector('.delete').addEventListener('click', async () => {
                await deleteComment(comment.id);
            });

            commentsContainer.appendChild(commentElement);
        });
        // 渲染完评论后再渲染分页按钮
        showPagination();
    }

    // 添加评论的函数
    async function addComment(name, content) {
        const response = await fetch(`${apiUrl}/add`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, content })
        });

        const result = await response.json();

        if (result.code === 0) {
            fetchComments(); // 重新获取评论并更新显示
        } else {
            alert('Failed to add comment: ' + result.msg);
        }
    }

    // 删除评论的函数
    async function deleteComment(id) {
        const response = await fetch(`${apiUrl}/delete?id=${id}`, {
            method: 'POST'
        });

        const result = await response.json();

        if (result.code === 0) {
            fetchComments(); // 重新获取评论并更新显示
        } else {
            alert('Failed to delete comment: ' + result.msg);
        }
    }

    // 实现分页功能，包括分页逻辑和分页按钮的渲染
    function showPagination() {
        const paginationContainer = document.createElement('div');
        paginationContainer.className = 'pagination';

        const totalPages = Math.ceil(totalComments / maxCommentsPerPage);

        const prevBtn = document.createElement('button');
        prevBtn.className = 'page-btn';
        prevBtn.innerText = '上一页';

        if (currentPage > 1) {
            prevBtn.addEventListener('click', () => {
                currentPage--;
                fetchComments();
            });
        } else {
            prevBtn.disabled = true;
        }

        paginationContainer.appendChild(prevBtn);

        const nextBtn = document.createElement('button');
        nextBtn.className = 'page-btn';
        nextBtn.innerText = '下一页';

        if (currentPage < totalPages) {
            nextBtn.addEventListener('click', () => {
                currentPage++;
                fetchComments();
            });
        } else {
            nextBtn.disabled = true;
        }

        paginationContainer.appendChild(nextBtn);
        commentsContainer.appendChild(paginationContainer);
    }

    // 提交评论主要逻辑，同时对用户名和评论内容进行一次检测
    submitBtn.addEventListener('click', async () => {
        const name = nameInput.value.trim();
        const comment = commentInput.value.trim();

        if (!name || !comment) {
            alert('请填写用户名和评论内容！');
            return;
        }
        // 给后端发完请求后清除输入框
        await addComment(name, comment);
        nameInput.value = '';
        commentInput.value = '';
    });

    // 初始化获取评论，在所有DOM元素加载完毕后第一个执行的语句
    fetchComments();
});
