<template>
  <div class="container">
    <h1>Daftar Peminjam Perpustakaan</h1>

    <input 
      type="text" 
      v-model="searchQuery" 
      placeholder="Cari nama atau no telepon..." 
      class="search-box"
    />

    <button class="add-btn" @click="showForm = true">
      <span class="plus-icon">+</span>
    </button>

    <div v-if="notification.message" :class="['notification', notification.type]">
      {{ notification.message }}
    </div>

    <div v-if="showForm" class="modal">
      <div class="modal-content">
        <h2>{{ isEditing ? 'Edit Peminjam' : 'Tambah Peminjam' }}</h2>
        <form @submit.prevent="savePeminjam">
          <input type="hidden" v-model="currentPeminjam.id" />
          <input type="text" v-model="currentPeminjam.user" placeholder="Nama" required />
          <input type="text" v-model="currentPeminjam.no_telepon" placeholder="No. Telepon" required />
          <input type="text" v-model="currentPeminjam.buku" placeholder="Buku" required />
          <input type="date" v-model="currentPeminjam.tanggal" required />
          <div class="form-actions">
            <button type="submit" class="btn save">{{ isEditing ? 'Update' : 'Simpan' }}</button>
            <button type="button" class="btn cancel" @click="closeForm">Batal</button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showConfirm" class="modal">
      <div class="modal-content confirm-box">
        <p>Yakin ingin menghapus peminjam ini?</p>
        <div class="confirm-actions">
          <button class="yes" @click="confirmDelete">Ya</button>
          <button class="no" @click="cancelDelete">Tidak</button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="loading">Sedang memuat data...</div>

    <table v-else border="1" cellpadding="10">
      <thead>
        <tr>
          <th>No</th>
          <th>Nama</th>
          <th>No. Telepon</th>
          <th>Buku</th>
          <th>Tanggal Pinjam</th>
          <th>Tanggal Kembali</th>
          <th>Aksi</th>
        </tr>
      </thead>
      <tbody>
        <tr 
          v-for="(peminjam, index) in filteredPeminjam" 
          :key="peminjam.id"
          :class="{ expiredRow: isExpired(peminjam.tanggal) }"
        >
          <td>{{ index + 1 }}</td>
          <td>{{ peminjam.user }}</td>
          <td>{{ peminjam.no_telepon }}</td>
          <td>{{ peminjam.buku }}</td>
          <td>{{ formatDate(peminjam.tanggal) }}</td>
          <td>{{ formatDate(hitungTanggalKembali(peminjam.tanggal)) }}</td>
          <td>
            <button class="btn edit" @click="editPeminjam(peminjam)">Edit</button>
            <button class="btn delete" @click="promptDelete(peminjam.id)">Hapus</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "Peminjam",
  data() {
    return {
      peminjamList: [],
      currentPeminjam: { id: null, user: '', no_telepon: '', buku: '', tanggal: '' },
      isEditing: false,
      showForm: false,
      notification: { message: '', type: '' },
      showConfirm: false,
      deleteId: null,
      loading: false,
      searchQuery: "" 
    };
  },
  mounted() {
    this.fetchPeminjam();
  },
  computed: {
    filteredPeminjam() {
      if (!this.searchQuery) return this.peminjamList;
      const q = this.searchQuery.toLowerCase();
      return this.peminjamList.filter(p =>
        p.user.toLowerCase().includes(q) ||
        p.no_telepon.toLowerCase().includes(q)
      );
    }
  },
  methods: {
    async fetchPeminjam() {
      this.loading = true;
      try {
        const response = await axios.get("http://localhost:8000/api/peminjam");
        this.peminjamList = response.data;
      } catch (error) {
        this.showNotification("Gagal mengambil data peminjam!", "error");
      } finally {
        this.loading = false;
      }
    },
    validateForm() {
      const { user, no_telepon, buku, tanggal } = this.currentPeminjam;

      if (!user || user.trim().length < 3) {
        this.showNotification("Nama minimal 3 huruf!", "error");
        return false;
      }
      if (!/^\d{10,}$/.test(no_telepon)) {
        this.showNotification("No. telepon harus angka minimal 10 digit!", "error");
        return false;
      }
      if (!buku || buku.trim().length < 2) {
        this.showNotification("Nama buku tidak boleh kosong!", "error");
        return false;
      }
      if (!tanggal) {
        this.showNotification("Tanggal pinjam wajib diisi!", "error");
        return false;
      }
      return true;
    },
    async savePeminjam() {
      if (!this.validateForm()) return;

      if (this.isEditing) {
        await this.updatePeminjam();
      } else {
        await this.createPeminjam();
      }
      this.resetForm();
    },
    async createPeminjam() {
      try {
        const response = await axios.post("http://localhost:8000/api/peminjam", this.currentPeminjam);
        this.peminjamList.push(response.data);
        this.showNotification("Peminjam berhasil ditambahkan!", "success");
      } catch (error) {
        this.showNotification("Gagal menambah peminjam!", "error");
      }
    },
    async updatePeminjam() {
      try {
        await axios.put(
          `http://localhost:8000/api/peminjam/${this.currentPeminjam.id}`,
          this.currentPeminjam
        );
        const index = this.peminjamList.findIndex(p => p.id === this.currentPeminjam.id);
        if (index !== -1) {
          this.peminjamList.splice(index, 1, { ...this.currentPeminjam });
        }
        this.showNotification("Data peminjam berhasil diupdate!", "success");
      } catch (error) {
        this.showNotification("Gagal mengupdate peminjam!", "error");
      }
    },
    promptDelete(id) {
      this.deleteId = id;
      this.showConfirm = true;
    },
    async confirmDelete() {
      try {
        await axios.delete(`http://localhost:8000/api/peminjam/${this.deleteId}`);
        this.peminjamList = this.peminjamList.filter(p => p.id !== this.deleteId);
        this.showNotification("Peminjam berhasil dihapus!", "success");
      } catch (error) {
        this.showNotification("Gagal menghapus peminjam!", "error");
      } finally {
        this.showConfirm = false;
        this.deleteId = null;
      }
    },
    cancelDelete() {
      this.showConfirm = false;
      this.deleteId = null;
    },
    editPeminjam(peminjam) {
      this.currentPeminjam = { ...peminjam };
      this.isEditing = true;
      this.showForm = true;
    },
    resetForm() {
      this.currentPeminjam = { id: null, user: '', no_telepon: '', buku: '', tanggal: '' };
      this.isEditing = false;
      this.showForm = false;
    },
    closeForm() {
      this.resetForm();
    },
    showNotification(message, type) {
      this.notification = { message, type };
      setTimeout(() => {
        this.notification = { message: '', type: '' };
      }, 3000);
    },
    hitungTanggalKembali(tanggalPinjam) {
      if (!tanggalPinjam) return null;
      const tgl = new Date(tanggalPinjam);
      tgl.setDate(tgl.getDate() + 7);
      return tgl;
    },
    isExpired(tanggalPinjam) {
      if (!tanggalPinjam) return false;
      const tglKembali = this.hitungTanggalKembali(tanggalPinjam);
      const today = new Date();
      return today > tglKembali;
    },
    formatDate(date) {
      if (!date) return "-";
      return new Date(date).toLocaleDateString("id-ID", {
        day: "2-digit",
        month: "long",
        year: "numeric"
      });
    }
  }
};
</script>

<style scoped>
.container {
  max-width: 900px;
  margin: 40px auto;
  font-family: Arial, sans-serif;
  color: #333;
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

h1 {
  text-align: center;
  margin-bottom: 20px;
  color: #2c3e50;
}

.search-box {
  width: 100%;
  padding: 8px;
  margin-bottom: 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.add-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background-color: #2c3e50;
  color: #fff;
  border: none;
  border-radius: 50%;
  font-size: 20px;
  cursor: pointer;
  margin-bottom: 15px;
}

.plus-icon {
  font-weight: bold;
  font-size: 24px;
}

.notification {
  text-align: center;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
  animation: fadeInOut 3s;
}
.notification.success { background-color: #4caf50; color: white; }
.notification.error { background-color: #e74c3c; color: white; }

@keyframes fadeInOut {
  0% { opacity: 0; }
  10% { opacity: 1; }
  90% { opacity: 1; }
  100% { opacity: 0; }
}

.loading {
  text-align: center;
  font-size: 18px;
  color: #2c3e50;
  margin: 20px 0;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}
.modal-content {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  width: 400px;
}

.confirm-box p {
  text-align: center;
  font-size: 16px;
  margin-bottom: 20px;
}
.confirm-actions {
  display: flex;
  justify-content: space-around;
}
.confirm-actions .yes {
  background: #e74c3c;
  color: white;
  border: none;
  padding: 8px 16px;
  cursor: pointer;
  border-radius: 4px;
}
.confirm-actions .no {
  background: #95a5a6;
  color: white;
  border: none;
  padding: 8px 16px;
  cursor: pointer;
  border-radius: 4px;
}

.form-actions {
  margin-top: 10px;
  display: flex;
  justify-content: space-between;
}
form input {
  display: block;
  width: 100%;
  margin-bottom: 10px;
  padding: 8px;
}

table {
  width: 100%;
  border-collapse: collapse;
}
th, td {
  border: 1px solid #ddd;
  padding: 12px;
  text-align: left;
}
th {
  background-color: #3498db;
  color: white;
}
tr:nth-child(even) { background: #f9f9f9; }


.expiredRow {
  background-color: #ffe6e6 !important;
  color: red;
  font-weight: bold;
}


.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
.btn.save { background: #27ae60; color: white; }
.btn.cancel { background: #95a5a6; color: white; }
.btn.edit { background: #2980b9; color: white; margin-right: 5px; }
.btn.delete { background: #e74c3c; color: white; }
</style>
