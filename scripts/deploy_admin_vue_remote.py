import base64
import os
import subprocess
from pathlib import Path

ROOT = Path('/root/blog-butterfly-go')
REMOTE = 'root@192.168.10.210'
PASSWORD = 'ywz0207.'
REMOTE_ROOT = '/tmp/blog-butterfly-go-deploy'
EXCLUDE = {'node_modules', 'dist', '.git', '.vscode'}


def run(cmd: str, timeout: int = 300):
    print(f'>>> {cmd}', flush=True)
    p = subprocess.run(cmd, shell=True, text=True, capture_output=True, timeout=timeout)
    if p.stdout:
        print(p.stdout, end='' if p.stdout.endswith('\n') else '\n', flush=True)
    if p.stderr:
        print(p.stderr, end='' if p.stderr.endswith('\n') else '\n', flush=True)
    if p.returncode != 0:
        raise SystemExit(p.returncode)


def ssh(cmd: str, timeout: int = 300):
    safe = cmd.replace("'", "'\\''")
    run(f"sshpass -p '{PASSWORD}' ssh -o StrictHostKeyChecking=no {REMOTE} '{safe}'", timeout=timeout)


paths = []
for src in [ROOT / 'admin-vue', ROOT / 'k8s' / 'admin-vue.yaml']:
    if src.is_file():
        paths.append(src)
    else:
        for p in src.rglob('*'):
            if p.is_dir():
                continue
            rel_parts = p.relative_to(ROOT).parts
            if any(part in EXCLUDE for part in rel_parts):
                continue
            paths.append(p)

paths = sorted(set(paths))
print(f'TOTAL_FILES={len(paths)}', flush=True)
ssh(f'rm -rf {REMOTE_ROOT} && mkdir -p {REMOTE_ROOT}/admin-vue {REMOTE_ROOT}/k8s')

for idx, p in enumerate(paths, 1):
    rel = p.relative_to(ROOT).as_posix()
    parent = os.path.dirname(f'{REMOTE_ROOT}/{rel}')
    ssh(f'mkdir -p {parent}')
    b64 = base64.b64encode(p.read_bytes()).decode('ascii')
    ssh(f"python3 -c \"import base64,pathlib; pathlib.Path('{REMOTE_ROOT}/{rel}').write_bytes(base64.b64decode('{b64}'))\"", timeout=300)
    print(f'[{idx}/{len(paths)}] synced {rel}', flush=True)

ssh(f'find {REMOTE_ROOT} -maxdepth 3 -type f | sort | sed -n "1,120p"', timeout=120)
ssh(f'set -e; cd {REMOTE_ROOT}/admin-vue; docker build -t blog-butterfly-admin-vue:latest .', timeout=1800)
ssh(f'set -e; cd {REMOTE_ROOT}; kubectl apply -f k8s/admin-vue.yaml; kubectl rollout restart deployment/blog-butterfly-admin-vue -n blog-butterfly-go; kubectl rollout status deployment/blog-butterfly-admin-vue -n blog-butterfly-go --timeout=300s; kubectl get pods -n blog-butterfly-go -l app=blog-butterfly-admin-vue -o wide; echo ---SVC---; kubectl get svc blog-butterfly-admin-vue -n blog-butterfly-go -o wide', timeout=600)
ssh("curl -fsS http://127.0.0.1:31085/ | sed -n '1,30p'", timeout=120)
