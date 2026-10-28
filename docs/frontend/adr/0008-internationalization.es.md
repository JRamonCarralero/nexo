# ADR-0008: Internacionalización con i18next y errores traducidos por código

[English](0008-internationalization.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

Nexo es bilingüe (español e inglés) desde el MVP. Los mensajes de error del backend están pensados para desarrolladores, no para el usuario final.

## Decisión

- **i18next** + **react-i18next**, con los textos en `shared/platform/i18n/locales/{es,en}.json`; el idioma se toma del navegador y el inglés es el de respaldo.
- Los errores se traducen **a partir del `code` del backend** (`errors.<CODE>`) con `useErrorMessage`, con un mensaje genérico de respaldo, y se muestra el `traceId` para que el usuario pueda comunicarlo.

## Consecuencias

**Positivas**

- Los mensajes están localizados y no dependen del texto del backend.

**Negativas / contrapartidas**

- Cada código nuevo del backend necesita su traducción en los dos idiomas.

## Alternativas consideradas

- **Lingui / FormatJS:** buenas opciones; i18next tiene el ecosistema más grande.
