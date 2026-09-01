const KEY = "wacalls.clientId";

const generate = (): string => {
  if (typeof crypto.randomUUID === "function") return crypto.randomUUID();
  return "c-" + Math.random().toString(36).slice(2) + Date.now().toString(36);
};

// Permite forcar a identidade do atendente pela URL: ?atendente=2
//
// Sem isso, so vale a identidade guardada no navegador — e duas janelas do
// mesmo navegador compartilham a mesma, o que faz o servidor ver uma pessoa
// so. Serve para testar varios atendentes sem precisar de outra maquina.
//
// Na integracao com o Morada Connect este valor passa a ser o usuario logado.
const fromURL = (): string | null => {
  try {
    const v = new URLSearchParams(window.location.search).get("atendente");
    return v && v.trim() !== "" ? `atendente-${v.trim()}` : null;
  } catch {
    return null;
  }
};

export const getClientId = (): string => {
  const forced = fromURL();
  if (forced) return forced;

  let id = localStorage.getItem(KEY);
  if (!id) {
    id = generate();
    localStorage.setItem(KEY, id);
  }
  return id;
};
