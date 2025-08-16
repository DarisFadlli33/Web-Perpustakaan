<template>
  <div>
    <Peminjam />
  </div>
</template>

<script>
import Peminjam from './components/Peminjam.vue';

export default {
  name: 'App',
  components: {
    Peminjam
  }
};
</script>

<style>
/* Gaya global dapat ditambahkan di sini jika diperlukan */
body {
  margin: 0;
  font-family: Arial, sans-serif;
  background-color: #f4f4f4;
}
</style>