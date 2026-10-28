# ADR-0001: Aplicación de una sola página con React y Vite

[English](0001-react-vite-spa.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

La aplicación vive detrás de un login, así que el SEO y el renderizado en servidor no importan; lo que importa es la rapidez de la interacción. Necesita librerías maduras para arrastrar y soltar, una paleta de comandos y componentes accesibles. El punto fuerte del autor es el backend; el frontend se construirá con mucha ayuda de herramientas e IA, así que el tamaño del ecosistema importa.

## Decisión

Una aplicación de una sola página en **React** construida con **Vite** y servida como ficheros estáticos. En desarrollo, Vite redirige `/api` a la API de Go, así que no hay CORS.

## Consecuencias

**Positivas**

- El ecosistema más rico para lo que necesita Nexo (dnd-kit, cmdk, Radix/shadcn).
- El alojamiento estático (Firebase Hosting o Cloudflare Pages) es gratuito.
- Los otros proyectos del autor ya usan React y Vite.

**Negativas / contrapartidas**

- La primera carga descarga toda la aplicación (lo mitiga la división de código por rutas).

## Alternativas consideradas

- **Angular:** el autor ya lo ha usado y se estructura como el código de backend, pero tiene menos librerías para tableros y paletas de comandos.
- **Next.js:** el renderizado en servidor no aporta nada detrás de un login y añade un servidor que mantener.
