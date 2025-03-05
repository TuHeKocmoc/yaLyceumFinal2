// web/static/main.js

function loadExpressions() {
    fetch('/api/v1/expressions') 
      .then(resp => {
        if (!resp.ok) {
          throw new Error('Failed to fetch expressions: ' + resp.status);
        }
        return resp.json();
      })
      .then(data => {
        // data = { expressions: [ {id, raw, status, result}, ... ] }
        const wrapper = document.getElementById('expressionsWrapper');
        if (!wrapper) return;
  
        const exprs = data.expressions || [];
        let html = '<ul>';
        exprs.forEach(e => {
          html += `<li>
            <strong>${e.id}</strong>:
            <em>${e.raw}</em> →
            Status: ${e.status}, Result: ${e.result !== null ? e.result : 'nil'}
          </li>`;
        });
        html += '</ul>';
  
        wrapper.innerHTML = html;
      })
      .catch(err => {
        console.error('Error loading expressions', err);
      });
  }
  
  setInterval(loadExpressions, 1000);
  
  loadExpressions();