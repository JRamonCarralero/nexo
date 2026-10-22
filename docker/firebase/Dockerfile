# Firebase Local Emulator Suite for development and tests.
# Only the Auth emulator for now (pure Node: no Java needed). Add a JRE when
# emulators that need it (Firestore) are enabled.
FROM node:24-alpine

ARG FIREBASE_TOOLS_VERSION=15.34.1
RUN npm install -g firebase-tools@${FIREBASE_TOOLS_VERSION} && npm cache clean --force

WORKDIR /firebase
COPY firebase.json .firebaserc ./

# 9099: Auth emulator · 4000: Emulator UI
EXPOSE 9099 4000
# "demo-" projects need no real Firebase project nor credentials.
CMD ["firebase", "emulators:start", "--project", "demo-nexo", "--only", "auth,ui"]
