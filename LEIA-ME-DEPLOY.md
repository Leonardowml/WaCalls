# Como subir no EasyPanel

Teste do WaCalls — ligações de WhatsApp pelo navegador.
Isso é um teste, separado do Morada Connect. Não encosta no sistema que está no ar.

## Antes de começar

- Tenha um **chip de teste**. Não use o número do condomínio.
- Escolha uma **senha** de acesso. Você vai digitar ela no painel, não aqui.

## Passo a passo

**1. Criar o serviço**
- No EasyPanel: **+ Service → App**
- Em *Source*, conecte este repositório (`Leonardowml/WaCalls`)
- Em *Build Method*, escolha **Dockerfile**
- Em *Dockerfile Path*, deixe `./Dockerfile`

**2. Colocar a senha**

Na aba *Environment*, adicione as duas linhas abaixo (troque pela senha que você escolheu):

```
WACALLS_USER=admin
WACALLS_PASSWORD=sua-senha-forte-aqui
```

Se esquecer isso, o programa não sobe de propósito — é a trava pra ninguém
expor o sistema sem senha por acidente.

**3. Guardar a sessão do WhatsApp**

Na aba *Volumes* (ou *Mounts*), crie um volume apontando para:

```
/data
```

Sem isso, toda vez que você atualizar vai ter que ler o QR Code de novo.

**4. Endereço**

Na aba *Domains*, use a porta **8080** e deixe o EasyPanel gerar o endereço
com cadeado (https). **Sempre acesse por esse endereço.**
Se entrar pelo número do servidor, o navegador bloqueia o microfone e parece
que está quebrado — mas é só isso.

**5. Instalar**

Clique em *Deploy*. Quando terminar, abra o endereço: vai pedir usuário e senha,
e depois mostrar a tela pra ler o QR Code com o chip de teste.

## Se a ligação completar mas ninguém ouvir

É a saída de áudio bloqueada na VPS. Peça ao suporte pra liberar
**tráfego UDP de saída**. É o primeiro suspeito, antes de achar que o programa falhou.

## O que foi alterado neste fork

- `cmd/server/auth.go` — trava de usuário e senha (o original vem sem nenhuma)
- `cmd/server/main.go` — duas linhas: ativa a trava e impede subir sem senha
- `Dockerfile`, `.dockerignore`, `docker-compose.yml` — para o EasyPanel montar

O resto é o código original, sem modificação.
